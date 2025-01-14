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

package etcdserver

import (
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"testing"

	"github.com/coreos/etcd/pkg/testutil"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/pkg/types"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/store"
)

func TestClusterMember(t *testing.T) {
	membs := []*Member{
		newTestMember(1, nil, "node1", nil),
		newTestMember(2, nil, "node2", nil),
	}
	tests := []struct {
		id    types.ID
		match bool
	}{
		{1, true},
		{2, true},
		{3, false},
	}
	for i, tt := range tests {
		c := newTestCluster(membs)
		m := c.Member(tt.id)
		if g := m != nil; g != tt.match {
			t.Errorf("#%d: find member = %v, want %v", i, g, tt.match)
		}
		if m != nil && m.ID != tt.id {
			t.Errorf("#%d: id = %x, want %x", i, m.ID, tt.id)
		}
	}
}

func TestClusterMemberByName(t *testing.T) {
	membs := []*Member{
		newTestMember(1, nil, "node1", nil),
		newTestMember(2, nil, "node2", nil),
	}
	tests := []struct {
		name  string
		match bool
	}{
		{"node1", true},
		{"node2", true},
		{"node3", false},
	}
	for i, tt := range tests {
		c := newTestCluster(membs)
		m := c.MemberByName(tt.name)
		if g := m != nil; g != tt.match {
			t.Errorf("#%d: find member = %v, want %v", i, g, tt.match)
		}
		if m != nil && m.Name != tt.name {
			t.Errorf("#%d: name = %v, want %v", i, m.Name, tt.name)
		}
	}
}

func TestClusterMemberIDs(t *testing.T) {
	c := newTestCluster([]*Member{
		newTestMember(1, nil, "", nil),
		newTestMember(4, nil, "", nil),
		newTestMember(100, nil, "", nil),
	})
	w := []types.ID{1, 4, 100}
	g := c.MemberIDs()
	if !reflect.DeepEqual(w, g) {
		t.Errorf("IDs = %+v, want %+v", g, w)
	}
}

func TestClusterPeerURLs(t *testing.T) {
	tests := []struct {
		mems  []*Member
		wurls []string
	}{
		// single peer with a single address
		{
			mems: []*Member{
				newTestMember(1, []string{"http://192.0.2.1"}, "", nil),
			},
			wurls: []string{"http://192.0.2.1"},
		},

		// single peer with a single address with a port
		{
			mems: []*Member{
				newTestMember(1, []string{"http://192.0.2.1:8001"}, "", nil),
			},
			wurls: []string{"http://192.0.2.1:8001"},
		},

		// several members explicitly unsorted
		{
			mems: []*Member{
				newTestMember(2, []string{"http://192.0.2.3", "http://192.0.2.4"}, "", nil),
				newTestMember(3, []string{"http://192.0.2.5", "http://192.0.2.6"}, "", nil),
				newTestMember(1, []string{"http://192.0.2.1", "http://192.0.2.2"}, "", nil),
			},
			wurls: []string{"http://192.0.2.1", "http://192.0.2.2", "http://192.0.2.3", "http://192.0.2.4", "http://192.0.2.5", "http://192.0.2.6"},
		},

		// no members
		{
			mems:  []*Member{},
			wurls: []string{},
		},

		// peer with no peer urls
		{
			mems: []*Member{
				newTestMember(3, []string{}, "", nil),
			},
			wurls: []string{},
		},
	}

	for i, tt := range tests {
		c := newTestCluster(tt.mems)
		urls := c.PeerURLs()
		if !reflect.DeepEqual(urls, tt.wurls) {
			t.Errorf("#%d: PeerURLs = %v, want %v", i, urls, tt.wurls)
		}
	}
}

func TestClusterClientURLs(t *testing.T) {
	tests := []struct {
		mems  []*Member
		wurls []string
	}{
		// single peer with a single address
		{
			mems: []*Member{
				newTestMember(1, nil, "", []string{"http://192.0.2.1"}),
			},
			wurls: []string{"http://192.0.2.1"},
		},

		// single peer with a single address with a port
		{
			mems: []*Member{
				newTestMember(1, nil, "", []string{"http://192.0.2.1:8001"}),
			},
			wurls: []string{"http://192.0.2.1:8001"},
		},

		// several members explicitly unsorted
		{
			mems: []*Member{
				newTestMember(2, nil, "", []string{"http://192.0.2.3", "http://192.0.2.4"}),
				newTestMember(3, nil, "", []string{"http://192.0.2.5", "http://192.0.2.6"}),
				newTestMember(1, nil, "", []string{"http://192.0.2.1", "http://192.0.2.2"}),
			},
			wurls: []string{"http://192.0.2.1", "http://192.0.2.2", "http://192.0.2.3", "http://192.0.2.4", "http://192.0.2.5", "http://192.0.2.6"},
		},

		// no members
		{
			mems:  []*Member{},
			wurls: []string{},
		},

		// peer with no client urls
		{
			mems: []*Member{
				newTestMember(3, nil, "", []string{}),
			},
			wurls: []string{},
		},
	}

	for i, tt := range tests {
		c := newTestCluster(tt.mems)
		urls := c.ClientURLs()
		if !reflect.DeepEqual(urls, tt.wurls) {
			t.Errorf("#%d: ClientURLs = %v, want %v", i, urls, tt.wurls)
		}
	}
}

func TestClusterValidateAndAssignIDsBad(t *testing.T) {
	tests := []struct {
		clmembs []*Member
		membs   []*Member
	}{
		{
			// unmatched length
			[]*Member{
				newTestMember(1, []string{"http://127.0.0.1:2379"}, "", nil),
			},
			[]*Member{},
		},
		{
			// unmatched peer urls
			[]*Member{
				newTestMember(1, []string{"http://127.0.0.1:2379"}, "", nil),
			},
			[]*Member{
				newTestMember(1, []string{"http://127.0.0.1:4001"}, "", nil),
			},
		},
		{
			// unmatched peer urls
			[]*Member{
				newTestMember(1, []string{"http://127.0.0.1:2379"}, "", nil),
				newTestMember(2, []string{"http://127.0.0.2:2379"}, "", nil),
			},
			[]*Member{
				newTestMember(1, []string{"http://127.0.0.1:2379"}, "", nil),
				newTestMember(2, []string{"http://127.0.0.2:4001"}, "", nil),
			},
		},
	}
	for i, tt := range tests {
		ecl := newTestCluster(tt.clmembs)
		lcl := newTestCluster(tt.membs)
		if err := ValidateClusterAndAssignIDs(lcl, ecl); err == nil {
			t.Errorf("#%d: unexpected update success", i)
		}
	}
}

func TestClusterValidateAndAssignIDs(t *testing.T) {
	tests := []struct {
		clmembs []*Member
		membs   []*Member
		wids    []types.ID
	}{
		{
			[]*Member{
				newTestMember(1, []string{"http://127.0.0.1:2379"}, "", nil),
				newTestMember(2, []string{"http://127.0.0.2:2379"}, "", nil),
			},
			[]*Member{
				newTestMember(3, []string{"http://127.0.0.1:2379"}, "", nil),
				newTestMember(4, []string{"http://127.0.0.2:2379"}, "", nil),
			},
			[]types.ID{3, 4},
		},
	}
	for i, tt := range tests {
		lcl := newTestCluster(tt.clmembs)
		ecl := newTestCluster(tt.membs)
		if err := ValidateClusterAndAssignIDs(lcl, ecl); err != nil {
			t.Errorf("#%d: unexpect update error: %v", i, err)
		}
		if !reflect.DeepEqual(lcl.MemberIDs(), tt.wids) {
			t.Errorf("#%d: ids = %v, want %v", i, lcl.MemberIDs(), tt.wids)
		}
	}
}

func TestClusterValidateConfigurationChange(t *testing.T) {
	cl := newCluster("")
	cl.SetStore(store.New())
	for i := 1; i <= 4; i++ {
		attr := RaftAttributes{PeerURLs: []string{fmt.Sprintf("http://127.0.0.1:%d", i)}}
		cl.AddMember(&Member{ID: types.ID(i), RaftAttributes: attr})
	}
	cl.RemoveMember(4)

	attr := RaftAttributes{PeerURLs: []string{fmt.Sprintf("http://127.0.0.1:%d", 1)}}
	ctx, err := json.Marshal(&Member{ID: types.ID(5), RaftAttributes: attr})
	if err != nil {
		t.Fatal(err)
	}

	attr = RaftAttributes{PeerURLs: []string{fmt.Sprintf("http://127.0.0.1:%d", 5)}}
	ctx5, err := json.Marshal(&Member{ID: types.ID(5), RaftAttributes: attr})
	if err != nil {
		t.Fatal(err)
	}

	attr = RaftAttributes{PeerURLs: []string{fmt.Sprintf("http://127.0.0.1:%d", 3)}}
	ctx2to3, err := json.Marshal(&Member{ID: types.ID(2), RaftAttributes: attr})
	if err != nil {
		t.Fatal(err)
	}

	attr = RaftAttributes{PeerURLs: []string{fmt.Sprintf("http://127.0.0.1:%d", 5)}}
	ctx2to5, err := json.Marshal(&Member{ID: types.ID(2), RaftAttributes: attr})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		cc   raftpb.ConfChange
		werr error
	}{
		{
			raftpb.ConfChange{
				Type:   raftpb.ConfChangeRemoveNode,
				NodeID: 3,
			},
			nil,
		},
		{
			raftpb.ConfChange{
				Type:   raftpb.ConfChangeAddNode,
				NodeID: 4,
			},
			ErrIDRemoved,
		},
		{
			raftpb.ConfChange{
				Type:   raftpb.ConfChangeRemoveNode,
				NodeID: 4,
			},
			ErrIDRemoved,
		},
		{
			raftpb.ConfChange{
				Type:   raftpb.ConfChangeAddNode,
				NodeID: 1,
			},
			ErrIDExists,
		},
		{
			raftpb.ConfChange{
				Type:    raftpb.ConfChangeAddNode,
				NodeID:  5,
				Context: ctx,
			},
			ErrPeerURLexists,
		},
		{
			raftpb.ConfChange{
				Type:   raftpb.ConfChangeRemoveNode,
				NodeID: 5,
			},
			ErrIDNotFound,
		},
		{
			raftpb.ConfChange{
				Type:    raftpb.ConfChangeAddNode,
				NodeID:  5,
				Context: ctx5,
			},
			nil,
		},
		{
			raftpb.ConfChange{
				Type:    raftpb.ConfChangeUpdateNode,
				NodeID:  5,
				Context: ctx,
			},
			ErrIDNotFound,
		},
		// try to change the peer url of 2 to the peer url of 3
		{
			raftpb.ConfChange{
				Type:    raftpb.ConfChangeUpdateNode,
				NodeID:  2,
				Context: ctx2to3,
			},
			ErrPeerURLexists,
		},
		{
			raftpb.ConfChange{
				Type:    raftpb.ConfChangeUpdateNode,
				NodeID:  2,
				Context: ctx2to5,
			},
			nil,
		},
	}
	for i, tt := range tests {
		err := cl.ValidateConfigurationChange(tt.cc)
		if err != tt.werr {
			t.Errorf("#%d: validateConfigurationChange error = %v, want %v", i, err, tt.werr)
		}
	}
}

func TestClusterGenID(t *testing.T) {
	cs := newTestCluster([]*Member{
		newTestMember(1, nil, "", nil),
		newTestMember(2, nil, "", nil),
	})

	cs.genID()
	if cs.ID() == 0 {
		t.Fatalf("cluster.ID = %v, want not 0", cs.ID())
	}
	previd := cs.ID()

	cs.SetStore(&storeRecorder{})
	cs.AddMember(newTestMember(3, nil, "", nil))
	cs.genID()
	if cs.ID() == previd {
		t.Fatalf("cluster.ID = %v, want not %v", cs.ID(), previd)
	}
}

func TestNodeToMemberBad(t *testing.T) {
	tests := []*store.NodeExtern{
		{Key: "/1234", Nodes: []*store.NodeExtern{
			{Key: "/1234/strange"},
		}},
		{Key: "/1234", Nodes: []*store.NodeExtern{
			{Key: "/1234/raftAttributes", Value: stringp("garbage")},
		}},
		{Key: "/1234", Nodes: []*store.NodeExtern{
			{Key: "/1234/attributes", Value: stringp(`{"name":"node1","clientURLs":null}`)},
		}},
		{Key: "/1234", Nodes: []*store.NodeExtern{
			{Key: "/1234/raftAttributes", Value: stringp(`{"peerURLs":null}`)},
			{Key: "/1234/strange"},
		}},
		{Key: "/1234", Nodes: []*store.NodeExtern{
			{Key: "/1234/raftAttributes", Value: stringp(`{"peerURLs":null}`)},
			{Key: "/1234/attributes", Value: stringp("garbage")},
		}},
		{Key: "/1234", Nodes: []*store.NodeExtern{
			{Key: "/1234/raftAttributes", Value: stringp(`{"peerURLs":null}`)},
			{Key: "/1234/attributes", Value: stringp(`{"name":"node1","clientURLs":null}`)},
			{Key: "/1234/strange"},
		}},
	}
	for i, tt := range tests {
		if _, err := nodeToMember(tt); err == nil {
			t.Errorf("#%d: unexpected nil error", i)
		}
	}
}

func TestClusterAddMember(t *testing.T) {
	st := &storeRecorder{}
	c := newTestCluster(nil)
	c.SetStore(st)
	c.AddMember(newTestMember(1, nil, "node1", nil))

	wactions := []testutil.Action{
		{
			Name: "Create",
			Params: []interface{}{
				path.Join(storeMembersPrefix, "1", "raftAttributes"),
				false,
				`{"peerURLs":null}`,
				false,
				store.Permanent,
			},
		},
	}
	if g := st.Action(); !reflect.DeepEqual(g, wactions) {
		t.Errorf("actions = %v, want %v", g, wactions)
	}
}

func TestClusterMembers(t *testing.T) {
	cls := &cluster{
		members: map[types.ID]*Member{
			1:   &Member{ID: 1},
			20:  &Member{ID: 20},
			100: &Member{ID: 100},
			5:   &Member{ID: 5},
			50:  &Member{ID: 50},
		},
	}
	w := []*Member{
		&Member{ID: 1},
		&Member{ID: 5},
		&Member{ID: 20},
		&Member{ID: 50},
		&Member{ID: 100},
	}
	if g := cls.Members(); !reflect.DeepEqual(g, w) {
		t.Fatalf("Members()=%#v, want %#v", g, w)
	}
}

func TestClusterRemoveMember(t *testing.T) {
	st := &storeRecorder{}
	c := newTestCluster(nil)
	c.SetStore(st)
	c.RemoveMember(1)

	wactions := []testutil.Action{
		{Name: "Delete", Params: []interface{}{memberStoreKey(1), true, true}},
		{Name: "Create", Params: []interface{}{removedMemberStoreKey(1), false, "", false, store.Permanent}},
	}
	if !reflect.DeepEqual(st.Action(), wactions) {
		t.Errorf("actions = %v, want %v", st.Action(), wactions)
	}
}

func TestClusterUpdateAttributes(t *testing.T) {
	name := "etcd"
	clientURLs := []string{"http://127.0.0.1:4001"}
	tests := []struct {
		mems    []*Member
		removed map[types.ID]bool
		wmems   []*Member
	}{
		// update attributes of existing member
		{
			[]*Member{
				newTestMember(1, nil, "", nil),
			},
			nil,
			[]*Member{
				newTestMember(1, nil, name, clientURLs),
			},
		},
		// update attributes of removed member
		{
			nil,
			map[types.ID]bool{types.ID(1): true},
			nil,
		},
	}
	for i, tt := range tests {
		c := newTestCluster(tt.mems)
		c.removed = tt.removed

		c.UpdateAttributes(types.ID(1), Attributes{Name: name, ClientURLs: clientURLs})
		if g := c.Members(); !reflect.DeepEqual(g, tt.wmems) {
			t.Errorf("#%d: members = %+v, want %+v", i, g, tt.wmems)
		}
	}
}

func TestNodeToMember(t *testing.T) {
	n := &store.NodeExtern{Key: "/1234", Nodes: []*store.NodeExtern{
		{Key: "/1234/attributes", Value: stringp(`{"name":"node1","clientURLs":null}`)},
		{Key: "/1234/raftAttributes", Value: stringp(`{"peerURLs":null}`)},
	}}
	wm := &Member{ID: 0x1234, RaftAttributes: RaftAttributes{}, Attributes: Attributes{Name: "node1"}}
	m, err := nodeToMember(n)
	if err != nil {
		t.Fatalf("unexpected nodeToMember error: %v", err)
	}
	if !reflect.DeepEqual(m, wm) {
		t.Errorf("member = %+v, want %+v", m, wm)
	}
}

func newTestCluster(membs []*Member) *cluster {
	c := &cluster{members: make(map[types.ID]*Member), removed: make(map[types.ID]bool)}
	for _, m := range membs {
		c.members[m.ID] = m
	}
	return c
}

func stringp(s string) *string { return &s }
