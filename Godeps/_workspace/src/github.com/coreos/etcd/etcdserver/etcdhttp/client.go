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

package etcdhttp

import (
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	etcdErr "github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/error"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/etcdserver"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/etcdserver/auth"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/etcdserver/etcdhttp/httptypes"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/etcdserver/stats"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/pkg/types"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/raft"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/store"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/coreos/etcd/version"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/jonboulle/clockwork"
	"github.com/coreos/mantle/Godeps/_workspace/src/github.com/prometheus/client_golang/prometheus"
	"github.com/coreos/mantle/Godeps/_workspace/src/golang.org/x/net/context"
)

const (
	authPrefix               = "/v2/auth"
	keysPrefix               = "/v2/keys"
	deprecatedMachinesPrefix = "/v2/machines"
	membersPrefix            = "/v2/members"
	statsPrefix              = "/v2/stats"
	varsPath                 = "/debug/vars"
	metricsPath              = "/metrics"
	healthPath               = "/health"
	versionPath              = "/version"
)

// NewClientHandler generates a muxed http.Handler with the given parameters to serve etcd client requests.
func NewClientHandler(server *etcdserver.EtcdServer) http.Handler {
	go capabilityLoop(server)

	sec := auth.NewStore(server, defaultServerTimeout)

	kh := &keysHandler{
		sec:     sec,
		server:  server,
		cluster: server.Cluster(),
		timer:   server,
		timeout: defaultServerTimeout,
	}

	sh := &statsHandler{
		stats: server,
	}

	mh := &membersHandler{
		sec:     sec,
		server:  server,
		cluster: server.Cluster(),
		clock:   clockwork.NewRealClock(),
	}

	dmh := &deprecatedMachinesHandler{
		cluster: server.Cluster(),
	}

	sech := &authHandler{
		sec:     sec,
		cluster: server.Cluster(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", http.NotFound)
	mux.Handle(healthPath, healthHandler(server))
	mux.HandleFunc(versionPath, versionHandler(server.Cluster(), serveVersion))
	mux.Handle(keysPrefix, kh)
	mux.Handle(keysPrefix+"/", kh)
	mux.HandleFunc(statsPrefix+"/store", sh.serveStore)
	mux.HandleFunc(statsPrefix+"/self", sh.serveSelf)
	mux.HandleFunc(statsPrefix+"/leader", sh.serveLeader)
	mux.HandleFunc(varsPath, serveVars)
	mux.Handle(metricsPath, prometheus.Handler())
	mux.Handle(membersPrefix, mh)
	mux.Handle(membersPrefix+"/", mh)
	mux.Handle(deprecatedMachinesPrefix, dmh)
	handleAuth(mux, sech)

	return requestLogger(mux)
}

type keysHandler struct {
	sec     *auth.Store
	server  etcdserver.Server
	cluster etcdserver.Cluster
	timer   etcdserver.RaftTimer
	timeout time.Duration
}

func (h *keysHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r.Method, "HEAD", "GET", "PUT", "POST", "DELETE") {
		return
	}

	w.Header().Set("X-Etcd-Cluster-ID", h.cluster.ID().String())

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	rr, err := parseKeyRequest(r, clockwork.NewRealClock())
	if err != nil {
		writeKeyError(w, err)
		return
	}
	// The path must be valid at this point (we've parsed the request successfully).
	if !hasKeyPrefixAccess(h.sec, r, r.URL.Path[len(keysPrefix):], rr.Recursive) {
		writeKeyNoAuth(w)
		return
	}

	resp, err := h.server.Do(ctx, rr)
	if err != nil {
		err = trimErrorPrefix(err, etcdserver.StoreKeysPrefix)
		writeKeyError(w, err)
		return
	}
	switch {
	case resp.Event != nil:
		if err := writeKeyEvent(w, resp.Event, h.timer); err != nil {
			// Should never be reached
			plog.Errorf("error writing event (%v)", err)
		}
	case resp.Watcher != nil:
		ctx, cancel := context.WithTimeout(context.Background(), defaultWatchTimeout)
		defer cancel()
		handleKeyWatch(ctx, w, resp.Watcher, rr.Stream, h.timer)
	default:
		writeKeyError(w, errors.New("received response with no Event/Watcher!"))
	}
}

type deprecatedMachinesHandler struct {
	cluster etcdserver.Cluster
}

func (h *deprecatedMachinesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r.Method, "GET", "HEAD") {
		return
	}
	endpoints := h.cluster.ClientURLs()
	w.Write([]byte(strings.Join(endpoints, ", ")))
}

type membersHandler struct {
	sec     *auth.Store
	server  etcdserver.Server
	cluster etcdserver.Cluster
	clock   clockwork.Clock
}

func (h *membersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r.Method, "GET", "POST", "DELETE", "PUT") {
		return
	}
	if !hasWriteRootAccess(h.sec, r) {
		writeNoAuth(w)
		return
	}
	w.Header().Set("X-Etcd-Cluster-ID", h.cluster.ID().String())

	ctx, cancel := context.WithTimeout(context.Background(), defaultServerTimeout)
	defer cancel()

	switch r.Method {
	case "GET":
		switch trimPrefix(r.URL.Path, membersPrefix) {
		case "":
			mc := newMemberCollection(h.cluster.Members())
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(mc); err != nil {
				plog.Warningf("failed to encode members response (%v)", err)
			}
		case "leader":
			id := h.server.Leader()
			if id == 0 {
				writeError(w, httptypes.NewHTTPError(http.StatusServiceUnavailable, "During election"))
				return
			}
			m := newMember(h.cluster.Member(id))
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(m); err != nil {
				plog.Warningf("failed to encode members response (%v)", err)
			}
		default:
			writeError(w, httptypes.NewHTTPError(http.StatusNotFound, "Not found"))
		}
	case "POST":
		req := httptypes.MemberCreateRequest{}
		if ok := unmarshalRequest(r, &req, w); !ok {
			return
		}
		now := h.clock.Now()
		m := etcdserver.NewMember("", req.PeerURLs, "", &now)
		err := h.server.AddMember(ctx, *m)
		switch {
		case err == etcdserver.ErrIDExists || err == etcdserver.ErrPeerURLexists:
			writeError(w, httptypes.NewHTTPError(http.StatusConflict, err.Error()))
			return
		case err != nil:
			plog.Errorf("error adding member %s (%v)", m.ID, err)
			writeError(w, err)
			return
		}
		res := newMember(m)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			plog.Warningf("failed to encode members response (%v)", err)
		}
	case "DELETE":
		id, ok := getID(r.URL.Path, w)
		if !ok {
			return
		}
		err := h.server.RemoveMember(ctx, uint64(id))
		switch {
		case err == etcdserver.ErrIDRemoved:
			writeError(w, httptypes.NewHTTPError(http.StatusGone, fmt.Sprintf("Member permanently removed: %s", id)))
		case err == etcdserver.ErrIDNotFound:
			writeError(w, httptypes.NewHTTPError(http.StatusNotFound, fmt.Sprintf("No such member: %s", id)))
		case err != nil:
			plog.Errorf("error removing member %s (%v)", id, err)
			writeError(w, err)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	case "PUT":
		id, ok := getID(r.URL.Path, w)
		if !ok {
			return
		}
		req := httptypes.MemberUpdateRequest{}
		if ok := unmarshalRequest(r, &req, w); !ok {
			return
		}
		m := etcdserver.Member{
			ID:             id,
			RaftAttributes: etcdserver.RaftAttributes{PeerURLs: req.PeerURLs.StringSlice()},
		}
		err := h.server.UpdateMember(ctx, m)
		switch {
		case err == etcdserver.ErrPeerURLexists:
			writeError(w, httptypes.NewHTTPError(http.StatusConflict, err.Error()))
		case err == etcdserver.ErrIDNotFound:
			writeError(w, httptypes.NewHTTPError(http.StatusNotFound, fmt.Sprintf("No such member: %s", id)))
		case err != nil:
			plog.Errorf("error updating member %s (%v)", m.ID, err)
			writeError(w, err)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

type statsHandler struct {
	stats stats.Stats
}

func (h *statsHandler) serveStore(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r.Method, "GET") {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(h.stats.StoreStats())
}

func (h *statsHandler) serveSelf(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r.Method, "GET") {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(h.stats.SelfStats())
}

func (h *statsHandler) serveLeader(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r.Method, "GET") {
		return
	}
	stats := h.stats.LeaderStats()
	if stats == nil {
		writeError(w, httptypes.NewHTTPError(http.StatusForbidden, "not current leader"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(stats)
}

func serveVars(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}

// TODO: change etcdserver to raft interface when we have it.
//       add test for healthHeadler when we have the interface ready.
func healthHandler(server *etcdserver.EtcdServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !allowMethod(w, r.Method, "GET") {
			return
		}

		if uint64(server.Leader()) == raft.None {
			http.Error(w, `{"health": "false"}`, http.StatusServiceUnavailable)
			return
		}

		// wait for raft's progress
		index := server.Index()
		for i := 0; i < 3; i++ {
			time.Sleep(250 * time.Millisecond)
			if server.Index() > index {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"health": "true"}`))
				return
			}
		}

		http.Error(w, `{"health": "false"}`, http.StatusServiceUnavailable)
		return
	}
}

func versionHandler(c etcdserver.Cluster, fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := c.Version()
		if v != nil {
			fn(w, r, v.String())
		} else {
			fn(w, r, "not_decided")
		}
	}
}

func serveVersion(w http.ResponseWriter, r *http.Request, clusterV string) {
	if !allowMethod(w, r.Method, "GET") {
		return
	}
	vs := version.Versions{
		Server:  version.Version,
		Cluster: clusterV,
	}

	b, err := json.Marshal(&vs)
	if err != nil {
		plog.Panicf("cannot marshal versions to json (%v)", err)
	}
	w.Write(b)
}

// parseKeyRequest converts a received http.Request on keysPrefix to
// a server Request, performing validation of supplied fields as appropriate.
// If any validation fails, an empty Request and non-nil error is returned.
func parseKeyRequest(r *http.Request, clock clockwork.Clock) (etcdserverpb.Request, error) {
	emptyReq := etcdserverpb.Request{}

	err := r.ParseForm()
	if err != nil {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodeInvalidForm,
			err.Error(),
		)
	}

	if !strings.HasPrefix(r.URL.Path, keysPrefix) {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodeInvalidForm,
			"incorrect key prefix",
		)
	}
	p := path.Join(etcdserver.StoreKeysPrefix, r.URL.Path[len(keysPrefix):])

	var pIdx, wIdx uint64
	if pIdx, err = getUint64(r.Form, "prevIndex"); err != nil {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodeIndexNaN,
			`invalid value for "prevIndex"`,
		)
	}
	if wIdx, err = getUint64(r.Form, "waitIndex"); err != nil {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodeIndexNaN,
			`invalid value for "waitIndex"`,
		)
	}

	var rec, sort, wait, dir, quorum, stream bool
	if rec, err = getBool(r.Form, "recursive"); err != nil {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodeInvalidField,
			`invalid value for "recursive"`,
		)
	}
	if sort, err = getBool(r.Form, "sorted"); err != nil {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodeInvalidField,
			`invalid value for "sorted"`,
		)
	}
	if wait, err = getBool(r.Form, "wait"); err != nil {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodeInvalidField,
			`invalid value for "wait"`,
		)
	}
	// TODO(jonboulle): define what parameters dir is/isn't compatible with?
	if dir, err = getBool(r.Form, "dir"); err != nil {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodeInvalidField,
			`invalid value for "dir"`,
		)
	}
	if quorum, err = getBool(r.Form, "quorum"); err != nil {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodeInvalidField,
			`invalid value for "quorum"`,
		)
	}
	if stream, err = getBool(r.Form, "stream"); err != nil {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodeInvalidField,
			`invalid value for "stream"`,
		)
	}

	if wait && r.Method != "GET" {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodeInvalidField,
			`"wait" can only be used with GET requests`,
		)
	}

	pV := r.FormValue("prevValue")
	if _, ok := r.Form["prevValue"]; ok && pV == "" {
		return emptyReq, etcdErr.NewRequestError(
			etcdErr.EcodePrevValueRequired,
			`"prevValue" cannot be empty`,
		)
	}

	// TTL is nullable, so leave it null if not specified
	// or an empty string
	var ttl *uint64
	if len(r.FormValue("ttl")) > 0 {
		i, err := getUint64(r.Form, "ttl")
		if err != nil {
			return emptyReq, etcdErr.NewRequestError(
				etcdErr.EcodeTTLNaN,
				`invalid value for "ttl"`,
			)
		}
		ttl = &i
	}

	// prevExist is nullable, so leave it null if not specified
	var pe *bool
	if _, ok := r.Form["prevExist"]; ok {
		bv, err := getBool(r.Form, "prevExist")
		if err != nil {
			return emptyReq, etcdErr.NewRequestError(
				etcdErr.EcodeInvalidField,
				"invalid value for prevExist",
			)
		}
		pe = &bv
	}

	rr := etcdserverpb.Request{
		Method:    r.Method,
		Path:      p,
		Val:       r.FormValue("value"),
		Dir:       dir,
		PrevValue: pV,
		PrevIndex: pIdx,
		PrevExist: pe,
		Wait:      wait,
		Since:     wIdx,
		Recursive: rec,
		Sorted:    sort,
		Quorum:    quorum,
		Stream:    stream,
	}

	if pe != nil {
		rr.PrevExist = pe
	}

	// Null TTL is equivalent to unset Expiration
	if ttl != nil {
		expr := time.Duration(*ttl) * time.Second
		rr.Expiration = clock.Now().Add(expr).UnixNano()
	}

	return rr, nil
}

// writeKeyEvent trims the prefix of key path in a single Event under
// StoreKeysPrefix, serializes it and writes the resulting JSON to the given
// ResponseWriter, along with the appropriate headers.
func writeKeyEvent(w http.ResponseWriter, ev *store.Event, rt etcdserver.RaftTimer) error {
	if ev == nil {
		return errors.New("cannot write empty Event!")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Etcd-Index", fmt.Sprint(ev.EtcdIndex))
	w.Header().Set("X-Raft-Index", fmt.Sprint(rt.Index()))
	w.Header().Set("X-Raft-Term", fmt.Sprint(rt.Term()))

	if ev.IsCreated() {
		w.WriteHeader(http.StatusCreated)
	}

	ev = trimEventPrefix(ev, etcdserver.StoreKeysPrefix)
	return json.NewEncoder(w).Encode(ev)
}

func writeKeyNoAuth(w http.ResponseWriter) {
	e := etcdErr.NewError(etcdErr.EcodeUnauthorized, "Insufficient credentials", 0)
	e.WriteTo(w)
}

// writeKeyError logs and writes the given Error to the ResponseWriter.
// If Error is not an etcdErr, the error will be converted to an etcd error.
func writeKeyError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	switch e := err.(type) {
	case *etcdErr.Error:
		e.WriteTo(w)
	default:
		if err == etcdserver.ErrTimeoutDueToLeaderFail {
			plog.Error(err)
		} else {
			plog.Errorf("got unexpected response error (%v)", err)
		}
		ee := etcdErr.NewError(etcdErr.EcodeRaftInternal, err.Error(), 0)
		ee.WriteTo(w)
	}
}

func handleKeyWatch(ctx context.Context, w http.ResponseWriter, wa store.Watcher, stream bool, rt etcdserver.RaftTimer) {
	defer wa.Remove()
	ech := wa.EventChan()
	var nch <-chan bool
	if x, ok := w.(http.CloseNotifier); ok {
		nch = x.CloseNotify()
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Etcd-Index", fmt.Sprint(wa.StartIndex()))
	w.Header().Set("X-Raft-Index", fmt.Sprint(rt.Index()))
	w.Header().Set("X-Raft-Term", fmt.Sprint(rt.Term()))
	w.WriteHeader(http.StatusOK)

	// Ensure headers are flushed early, in case of long polling
	w.(http.Flusher).Flush()

	for {
		select {
		case <-nch:
			// Client closed connection. Nothing to do.
			return
		case <-ctx.Done():
			// Timed out. net/http will close the connection for us, so nothing to do.
			return
		case ev, ok := <-ech:
			if !ok {
				// If the channel is closed this may be an indication of
				// that notifications are much more than we are able to
				// send to the client in time. Then we simply end streaming.
				return
			}
			ev = trimEventPrefix(ev, etcdserver.StoreKeysPrefix)
			if err := json.NewEncoder(w).Encode(ev); err != nil {
				// Should never be reached
				plog.Warningf("error writing event (%v)", err)
				return
			}
			if !stream {
				return
			}
			w.(http.Flusher).Flush()
		}
	}
}

func trimEventPrefix(ev *store.Event, prefix string) *store.Event {
	if ev == nil {
		return nil
	}
	// Since the *Event may reference one in the store history
	// history, we must copy it before modifying
	e := ev.Clone()
	e.Node = trimNodeExternPrefix(e.Node, prefix)
	e.PrevNode = trimNodeExternPrefix(e.PrevNode, prefix)
	return e
}

func trimNodeExternPrefix(n *store.NodeExtern, prefix string) *store.NodeExtern {
	if n == nil {
		return nil
	}
	n.Key = strings.TrimPrefix(n.Key, prefix)
	for _, nn := range n.Nodes {
		nn = trimNodeExternPrefix(nn, prefix)
	}
	return n
}

func trimErrorPrefix(err error, prefix string) error {
	if e, ok := err.(*etcdErr.Error); ok {
		e.Cause = strings.TrimPrefix(e.Cause, prefix)
	}
	return err
}

func unmarshalRequest(r *http.Request, req json.Unmarshaler, w http.ResponseWriter) bool {
	ctype := r.Header.Get("Content-Type")
	if ctype != "application/json" {
		writeError(w, httptypes.NewHTTPError(http.StatusUnsupportedMediaType, fmt.Sprintf("Bad Content-Type %s, accept application/json", ctype)))
		return false
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, httptypes.NewHTTPError(http.StatusBadRequest, err.Error()))
		return false
	}
	if err := req.UnmarshalJSON(b); err != nil {
		writeError(w, httptypes.NewHTTPError(http.StatusBadRequest, err.Error()))
		return false
	}
	return true
}

func getID(p string, w http.ResponseWriter) (types.ID, bool) {
	idStr := trimPrefix(p, membersPrefix)
	if idStr == "" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return 0, false
	}
	id, err := types.IDFromString(idStr)
	if err != nil {
		writeError(w, httptypes.NewHTTPError(http.StatusNotFound, fmt.Sprintf("No such member: %s", idStr)))
		return 0, false
	}
	return id, true
}

// getUint64 extracts a uint64 by the given key from a Form. If the key does
// not exist in the form, 0 is returned. If the key exists but the value is
// badly formed, an error is returned. If multiple values are present only the
// first is considered.
func getUint64(form url.Values, key string) (i uint64, err error) {
	if vals, ok := form[key]; ok {
		i, err = strconv.ParseUint(vals[0], 10, 64)
	}
	return
}

// getBool extracts a bool by the given key from a Form. If the key does not
// exist in the form, false is returned. If the key exists but the value is
// badly formed, an error is returned. If multiple values are present only the
// first is considered.
func getBool(form url.Values, key string) (b bool, err error) {
	if vals, ok := form[key]; ok {
		b, err = strconv.ParseBool(vals[0])
	}
	return
}

// trimPrefix removes a given prefix and any slash following the prefix
// e.g.: trimPrefix("foo", "foo") == trimPrefix("foo/", "foo") == ""
func trimPrefix(p, prefix string) (s string) {
	s = strings.TrimPrefix(p, prefix)
	s = strings.TrimPrefix(s, "/")
	return
}

func newMemberCollection(ms []*etcdserver.Member) *httptypes.MemberCollection {
	c := httptypes.MemberCollection(make([]httptypes.Member, len(ms)))

	for i, m := range ms {
		c[i] = newMember(m)
	}

	return &c
}

func newMember(m *etcdserver.Member) httptypes.Member {
	tm := httptypes.Member{
		ID:         m.ID.String(),
		Name:       m.Name,
		PeerURLs:   make([]string, len(m.PeerURLs)),
		ClientURLs: make([]string, len(m.ClientURLs)),
	}

	copy(tm.PeerURLs, m.PeerURLs)
	copy(tm.ClientURLs, m.ClientURLs)

	return tm
}
