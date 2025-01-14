// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package platform

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/ec2"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/mantle/Godeps/_workspace/src/golang.org/x/crypto/ssh"

	"github.com/coreos/mantle/network"
	"github.com/coreos/mantle/util"
)

type awsMachine struct {
	cluster   *awsCluster
	mach      *ec2.Instance
	sshClient *ssh.Client
}

func (am *awsMachine) ID() string {
	return *am.mach.InstanceId
}

func (am *awsMachine) IP() string {
	return *am.mach.PublicIpAddress
}

func (am *awsMachine) PrivateIP() string {
	return *am.mach.PrivateIpAddress
}

func (am *awsMachine) SSHSession() (*ssh.Session, error) {
	session, err := am.sshClient.NewSession()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (am *awsMachine) SSH(cmd string) ([]byte, error) {
	session, err := am.SSHSession()
	if err != nil {
		return []byte{}, err
	}
	defer session.Close()

	session.Stderr = os.Stderr
	out, err := session.Output(cmd)
	out = bytes.TrimSpace(out)
	return out, err
}

func (am *awsMachine) Destroy() error {
	if am.sshClient != nil {
		am.sshClient.Close()
	}

	id := am.ID()

	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{&id},
	}

	if _, err := am.cluster.api.TerminateInstances(input); err != nil {
		return err
	}

	am.cluster.delMach(am)
	return nil
}

func (am *awsMachine) StartJournal() error {
	s, err := am.SSHSession()
	if err != nil {
		return fmt.Errorf("SSH session failed: %v", err)
	}

	s.Stdout = os.Stdout
	s.Stderr = os.Stderr
	go func() {
		s.Run("journalctl -f")
		s.Close()
	}()

	return nil
}

type AWSOptions struct {
	AMI           string
	KeyName       string
	InstanceType  string
	SecurityGroup string
}
type awsCluster struct {
	mu    sync.Mutex
	api   *ec2.EC2
	conf  AWSOptions
	agent *network.SSHAgent
	machs map[string]*awsMachine
}

func NewAWSCluster(conf AWSOptions) (Cluster, error) {
	api := ec2.New(aws.NewConfig().WithCredentials(credentials.NewEnvCredentials()))

	agent, err := network.NewSSHAgent(&net.Dialer{})
	if err != nil {
		return nil, err
	}

	ac := &awsCluster{
		api:   api,
		conf:  conf,
		agent: agent,
		machs: make(map[string]*awsMachine),
	}

	return ac, nil
}

func (ac *awsCluster) addMach(m *awsMachine) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.machs[*m.mach.InstanceId] = m
}

func (ac *awsCluster) delMach(m *awsMachine) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	delete(ac.machs, *m.mach.InstanceId)
}

func (ac *awsCluster) NewCommand(name string, arg ...string) util.Cmd {
	return util.NewCommand(name, arg...)
}

func (ac *awsCluster) NewMachine(userdata string) (Machine, error) {
	cloudConfig, err := config.NewCloudConfig(userdata)
	if err != nil {
		return nil, err
	}

	if err = ac.agent.UpdateConfig(cloudConfig); err != nil {
		return nil, err
	}

	if cloudConfig.Hostname == "" {
		id := make([]byte, 4)
		_, _ = rand.Read(id)
		cloudConfig.Hostname = fmt.Sprintf("%x", id)
	}

	ud := base64.StdEncoding.EncodeToString([]byte(cloudConfig.String()))
	cnt := int64(1)

	inst := ec2.RunInstancesInput{
		ImageId:        &ac.conf.AMI,
		MinCount:       &cnt,
		MaxCount:       &cnt,
		KeyName:        &ac.conf.KeyName, // this is only useful if you wish to ssh in for debugging
		InstanceType:   &ac.conf.InstanceType,
		SecurityGroups: []*string{&ac.conf.SecurityGroup},
		UserData:       &ud,
	}

	resp, err := ac.api.RunInstances(&inst)
	if err != nil {
		return nil, err
	}

	ids := []*string{resp.Instances[0].InstanceId}

	if err := waitForAWSInstances(ac.api, ids, 5*time.Minute); err != nil {
		return nil, err
	}

	getinst := &ec2.DescribeInstancesInput{
		InstanceIds: ids,
	}

	insts, err := ac.api.DescribeInstances(getinst)
	if err != nil {
		return nil, err
	}

	mach := &awsMachine{
		cluster: ac,
		mach:    insts.Reservations[0].Instances[0],
	}

	// Allow a few authentication failures in case setup is slow.
	sshchecker := func() error {
		mach.sshClient, err = mach.cluster.agent.NewClient(mach.IP())
		if err != nil {
			return err
		}
		return nil
	}

	if err := util.Retry(sshRetries, sshTimeout, sshchecker); err != nil {
		mach.Destroy()
		return nil, err
	}

	ac.addMach(mach)

	return mach, nil
}

func (ac *awsCluster) Machines() []Machine {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	machs := make([]Machine, 0, len(ac.machs))
	for _, m := range ac.machs {
		machs = append(machs, m)
	}
	return machs
}

func (ac *awsCluster) EtcdEndpoint() string {
	return ""
}

func (ac *awsCluster) GetDiscoveryURL(size int) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://discovery.etcd.io/new?size=%d", size))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (ac *awsCluster) Destroy() error {
	machs := ac.Machines()
	for _, am := range machs {
		am.Destroy()
	}
	ac.agent.Close()
	return nil
}

// waitForAWSInstance waits until a set of aws ec2 instance is accessible by ssh.
func waitForAWSInstances(api *ec2.EC2, ids []*string, d time.Duration) error {
	after := time.After(d)

	online := make(map[string]bool)

	for len(ids) != len(online) {
		select {
		case <-after:
			return fmt.Errorf("timed out waiting for instances to run")
		default:
		}

		// don't make api calls too quickly, or we will hit the rate limit

		time.Sleep(10 * time.Second)

		getinst := &ec2.DescribeInstancesInput{
			InstanceIds: ids,
		}

		insts, err := api.DescribeInstances(getinst)
		if err != nil {
			return err
		}

		for _, r := range insts.Reservations {
			for _, i := range r.Instances {
				// skip instances known to be up
				if online[*i.InstanceId] {
					continue
				}

				// "running"
				if *i.State.Code == int64(16) {
					// XXX: ssh is a terrible way to check this, but it is all we have.
					c, err := net.DialTimeout("tcp", *i.PublicIpAddress+":22", 10*time.Second)
					if err != nil {
						continue
					}
					c.Close()

					online[*i.InstanceId] = true
				}
			}
		}
	}

	return nil
}
