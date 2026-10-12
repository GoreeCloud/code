// Copyright 2025 The Forgejo Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"forgejo.org/modules/util"

	ap "github.com/go-ap/activitypub"
	"github.com/go-ap/jsonld"
	"github.com/google/uuid"
)

type ApActorMock struct {
	PrivKey string
	PubKey  string
}

type FederationServerMockPerson struct {
	ID      int64
	Name    string
	PubKey  string
	PrivKey string
}

type FederationServerMockRepository struct {
	ID int64
}

type FederationServerMock struct {
	ApActor      ApActorMock
	Persons      []FederationServerMockPerson
	Repositories []FederationServerMockRepository
	LastPost     string
}

func NewApActorMock() ApActorMock {
	priv, pub, _ := util.GenerateKeyPair(1024)
	return ApActorMock{
		PrivKey: priv,
		PubKey:  pub,
	}
}

func (u *ApActorMock) KeyID(host string) string {
	return fmt.Sprintf("%s/api/v1/activitypub/actor#main-key", host)
}

func (u *ApActorMock) marshal(host string) string {
	baseID := fmt.Sprintf("http://%s/api/v1/activitypub/actor", host)

	return fmt.Sprintf(
		`{ "@context": ["https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"],`+
			`"id": "%[1]s",`+
			`"type": "Application",`+
			`"preferredUsername": "ghost",`+
			`"publicKey": {`+
			` "id": "%[1]s#main-key",`+
			` "owner": "%[1]s",`+
			` "publicKeyPem": %[2]q }}`,
		baseID,
		u.PubKey,
	)
}

func NewFederationServerMockPerson(id int64, name string) FederationServerMockPerson {
	priv, pub, _ := util.GenerateKeyPair(3072)
	return FederationServerMockPerson{
		ID:      id,
		Name:    name,
		PubKey:  pub,
		PrivKey: priv,
	}
}

func (p *FederationServerMockPerson) KeyID(host string) string {
	return fmt.Sprintf("%[1]v/api/v1/activitypub/user-id/%[2]v#main-key", host, p.ID)
}

func (p FederationServerMockPerson) marshal(host string) string {
	return fmt.Sprintf(`{"@context":["https://www.w3.org/ns/activitystreams","https://w3id.org/security/v1"],`+
		`"id":"http://%[1]v/api/v1/activitypub/user-id/%[2]v",`+
		`"type":"Person",`+
		`"icon":{"type":"Image","mediaType":"image/png","url":"http://%[1]v/avatars/1bb05d9a5f6675ed0272af9ea193063c"},`+
		`"url":"http://%[1]v/%[2]v",`+
		`"inbox":"http://%[1]v/api/v1/activitypub/user-id/%[2]v/inbox",`+
		`"outbox":"http://%[1]v/api/v1/activitypub/user-id/%[2]v/outbox",`+
		`"preferredUsername":"%[3]v",`+
		`"publicKey":{"id":"http://%[1]v/api/v1/activitypub/user-id/%[2]v#main-key",`+
		`"owner":"http://%[1]v/api/v1/activitypub/user-id/%[2]v",`+
		`"publicKeyPem":%[4]q}}`, host, p.ID, p.Name, p.PubKey)
}

func NewFederationServerMockRepository(id int64) FederationServerMockRepository {
	return FederationServerMockRepository{
		ID: id,
	}
}

func NewFederationServerMock() *FederationServerMock {
	return &FederationServerMock{
		ApActor: NewApActorMock(),
		Persons: []FederationServerMockPerson{
			NewFederationServerMockPerson(15, "stargoose1"),
			NewFederationServerMockPerson(30, "stargoose2"),
		},
		Repositories: []FederationServerMockRepository{
			NewFederationServerMockRepository(1),
		},
		LastPost: "",
	}
}

func (mock *FederationServerMock) recordLastPost(t *testing.T, req *http.Request) {
	buf := new(strings.Builder)
	_, err := io.Copy(buf, req.Body)
	if err != nil {
		t.Errorf("Error reading body: %q", err)
	}
	mock.LastPost = strings.ReplaceAll(buf.String(), req.Host, "DISTANT_FEDERATION_HOST")
}

func (mock *FederationServerMock) FollowActorUnsigned(host string, localID int64, uri, inboxURL url.URL) error {
	apID := fmt.Sprintf("%s/api/v1/activitypub/user-id/%d", host, localID)

	activity := ap.Follow{}
	activity.Type = ap.FollowType
	activity.ID = ap.IRI(apID + "/follows/" + uuid.New().String())
	activity.Actor = ap.IRI(apID)
	activity.Object = ap.IRI(uri.String())

	payload, err := jsonld.WithContext(jsonld.IRI(ap.ActivityBaseURI)).Marshal(activity)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(payload)
	_, err = http.Post(inboxURL.String(), "application/activity+json", reader)

	return err
}

func (mock *FederationServerMock) DistantServer(t *testing.T) *httptest.Server {
	federatedRoutes := http.NewServeMux()

	federatedRoutes.HandleFunc("/.well-known/nodeinfo",
		func(res http.ResponseWriter, req *http.Request) {
			// curl -H "Accept: application/json" https://federated-repo.prod.meissa.de/.well-known/nodeinfo
			// TODO: as soon as content-type will become important:  content-type: application/json;charset=utf-8
			fmt.Fprintf(res, `{"links":[{"href":"http://%s/api/v1/nodeinfo","rel":"http://nodeinfo.diaspora.software/ns/schema/2.1"}]}`, req.Host)
		})
	federatedRoutes.HandleFunc("/api/v1/nodeinfo",
		func(res http.ResponseWriter, req *http.Request) {
			// curl -H "Accept: application/json" https://federated-repo.prod.meissa.de/api/v1/nodeinfo
			fmt.Fprint(res, `{"version":"2.1","software":{"name":"forgejo","version":"1.20.0+dev-3183-g976d79044",`+
				`"repository":"https://codeberg.org/forgejo/forgejo.git","homepage":"https://forgejo.org/"},`+
				`"protocols":["activitypub"],"services":{"inbound":[],"outbound":["rss2.0"]},`+
				`"openRegistrations":true,"usage":{"users":{"total":14,"activeHalfyear":2}},"metadata":{}}`)
		})

	for _, person := range mock.Persons {
		federatedRoutes.HandleFunc(fmt.Sprintf("/api/v1/activitypub/user-id/alias%v", person.ID),
			func(res http.ResponseWriter, req *http.Request) {
				fmt.Fprint(res, person.marshal(req.Host))
			})
		federatedRoutes.HandleFunc(fmt.Sprintf("/api/v1/activitypub/user-id/%v", person.ID),
			func(res http.ResponseWriter, req *http.Request) {
				// curl -H "Accept: application/json" https://federated-repo.prod.meissa.de/api/v1/activitypub/user-id/2
				fmt.Fprint(res, person.marshal(req.Host))
			})
		federatedRoutes.HandleFunc(fmt.Sprintf("POST /api/v1/activitypub/user-id/%v/inbox", person.ID),
			func(res http.ResponseWriter, req *http.Request) {
				mock.recordLastPost(t, req)
			})
	}

	for _, repository := range mock.Repositories {
		federatedRoutes.HandleFunc(fmt.Sprintf("POST /api/v1/activitypub/repository-id/%v/inbox", repository.ID),
			func(res http.ResponseWriter, req *http.Request) {
				mock.recordLastPost(t, req)
			})
	}

	federatedRoutes.HandleFunc("GET /api/v1/activitypub/actor",
		func(res http.ResponseWriter, req *http.Request) {
			fmt.Fprint(res, mock.ApActor.marshal(req.Host))
		})

	federatedRoutes.HandleFunc("/",
		func(res http.ResponseWriter, req *http.Request) {
			t.Errorf("Unhandled %v request: %q", req.Method, req.URL.EscapedPath())
		})

	federatedSrv := httptest.NewServer(federatedRoutes)

	return federatedSrv
}
