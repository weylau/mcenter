package amqp

import (
	"testing"
)

// Test matrix defined on http://www.rabbitmq.com/uri-spec.html
type testURI struct {
	url      string
	username string
	password string
	host     string
	port     int
	vhost    string
	canon    string
}

var uriTests = []testURI{
	{
		url:      "queue://user:pass@host:10000/vhost",
		username: "user",
		password: "pass",
		host:     "host",
		port:     10000,
		vhost:    "vhost",
		canon:    "queue://user:pass@host:10000/vhost",
	},

	{
		url:      "queue://",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     defaultURI.Host,
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
		canon:    "queue://localhost/",
	},

	{
		url:      "queue://:@/",
		username: "",
		password: "",
		host:     defaultURI.Host,
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
		canon:    "queue://:@localhost/",
	},

	{
		url:      "queue://user@",
		username: "user",
		password: defaultURI.Password,
		host:     defaultURI.Host,
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
		canon:    "queue://user@localhost/",
	},

	{
		url:      "queue://user:pass@",
		username: "user",
		password: "pass",
		host:     defaultURI.Host,
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
		canon:    "queue://user:pass@localhost/",
	},

	{
		url:      "queue://guest:pass@",
		username: "guest",
		password: "pass",
		host:     defaultURI.Host,
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
		canon:    "queue://guest:pass@localhost/",
	},

	{
		url:      "queue://host",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "host",
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
		canon:    "queue://host/",
	},

	{
		url:      "queue://:10000",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     defaultURI.Host,
		port:     10000,
		vhost:    defaultURI.Vhost,
		canon:    "queue://localhost:10000/",
	},

	{
		url:      "queue:///vhost",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     defaultURI.Host,
		port:     defaultURI.Port,
		vhost:    "vhost",
		canon:    "queue://localhost/vhost",
	},

	{
		url:      "queue://host/",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "host",
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
		canon:    "queue://host/",
	},

	{
		url:      "queue://host/%2F",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "host",
		port:     defaultURI.Port,
		vhost:    "/",
		canon:    "queue://host/",
	},

	{
		url:      "queue://host/%2F%2F",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "host",
		port:     defaultURI.Port,
		vhost:    "//",
		canon:    "queue://host/%2F%2F",
	},

	{
		url:      "queue://host/%2Fslash%2F",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "host",
		port:     defaultURI.Port,
		vhost:    "/slash/",
		canon:    "queue://host/%2Fslash%2F",
	},

	{
		url:      "queue://192.168.1.1:1000/",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "192.168.1.1",
		port:     1000,
		vhost:    defaultURI.Vhost,
		canon:    "queue://192.168.1.1:1000/",
	},

	{
		url:      "queue://[::1]",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "::1",
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
		canon:    "queue://[::1]/",
	},

	{
		url:      "queue://[::1]:1000",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "::1",
		port:     1000,
		vhost:    defaultURI.Vhost,
		canon:    "queue://[::1]:1000/",
	},

	{
		url:      "queue://[fe80::1]",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "fe80::1",
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
		canon:    "queue://[fe80::1]/",
	},

	{
		url:      "queue://[fe80::1]",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "fe80::1",
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
		canon:    "queue://[fe80::1]/",
	},

	{
		url:      "queue://[fe80::1%25en0]",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "fe80::1%en0",
		port:     defaultURI.Port,
		vhost:    defaultURI.Vhost,
		canon:    "queue://[fe80::1%25en0]/",
	},

	{
		url:      "queue://[fe80::1]:5671",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "fe80::1",
		port:     5671,
		vhost:    defaultURI.Vhost,
		canon:    "queue://[fe80::1]:5671/",
	},

	{
		url:      "amqps:///",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     defaultURI.Host,
		port:     schemePorts["amqps"],
		vhost:    defaultURI.Vhost,
		canon:    "amqps://localhost/",
	},

	{
		url:      "amqps://host:1000/",
		username: defaultURI.Username,
		password: defaultURI.Password,
		host:     "host",
		port:     1000,
		vhost:    defaultURI.Vhost,
		canon:    "amqps://host:1000/",
	},
}

func TestURISpec(t *testing.T) {
	for _, test := range uriTests {
		u, err := ParseURI(test.url)
		if err != nil {
			t.Fatal("Could not parse spec URI: ", test.url, " err: ", err)
		}

		if test.username != u.Username {
			t.Error("For: ", test.url, " usernames do not match. want: ", test.username, " got: ", u.Username)
		}

		if test.password != u.Password {
			t.Error("For: ", test.url, " passwords do not match. want: ", test.password, " got: ", u.Password)
		}

		if test.host != u.Host {
			t.Error("For: ", test.url, " hosts do not match. want: ", test.host, " got: ", u.Host)
		}

		if test.port != u.Port {
			t.Error("For: ", test.url, " ports do not match. want: ", test.port, " got: ", u.Port)
		}

		if test.vhost != u.Vhost {
			t.Error("For: ", test.url, " vhosts do not match. want: ", test.vhost, " got: ", u.Vhost)
		}

		if test.canon != u.String() {
			t.Error("For: ", test.url, " canonical string does not match. want: ", test.canon, " got: ", u.String())
		}
	}
}

func TestURIUnknownScheme(t *testing.T) {
	if _, err := ParseURI("http://example.com/"); err == nil {
		t.Fatal("Expected error when parsing non-queue scheme")
	}
}

func TestURIScheme(t *testing.T) {
	if _, err := ParseURI("queue://example.com/"); err != nil {
		t.Fatalf("Expected to parse queue scheme, got %v", err)
	}

	if _, err := ParseURI("amqps://example.com/"); err != nil {
		t.Fatalf("Expected to parse amqps scheme, got %v", err)
	}
}

func TestURIWhitespace(t *testing.T) {
	if _, err := ParseURI("queue://admin:PASSWORD@rabbitmq-service/ -http_port=8080"); err == nil {
		t.Fatal("Expected to fail if URI contains whitespace")
	}
}

func TestURIDefaults(t *testing.T) {
	url := "queue://"
	uri, err := ParseURI(url)
	if err != nil {
		t.Fatal("Could not parse")
	}

	if uri.String() != "queue://localhost/" {
		t.Fatal("Defaults not encoded properly got:", uri.String())
	}
}

func TestURIComplete(t *testing.T) {
	url := "queue://bob:dobbs@foo.bar:5678/private"
	uri, err := ParseURI(url)
	if err != nil {
		t.Fatal("Could not parse")
	}

	if uri.String() != url {
		t.Fatal("Defaults not encoded properly want:", url, " got:", uri.String())
	}
}

func TestURIDefaultPortAmqpNotIncluded(t *testing.T) {
	url := "queue://foo.bar:5672/"
	uri, err := ParseURI(url)
	if err != nil {
		t.Fatal("Could not parse")
	}

	if uri.String() != "queue://foo.bar/" {
		t.Fatal("Defaults not encoded properly got:", uri.String())
	}
}

func TestURIDefaultPortAmqp(t *testing.T) {
	url := "queue://foo.bar/"
	uri, err := ParseURI(url)
	if err != nil {
		t.Fatal("Could not parse")
	}

	if uri.Port != 5672 {
		t.Fatal("Default port not correct for queue, got:", uri.Port)
	}
}

func TestURIDefaultPortAmqpsNotIncludedInString(t *testing.T) {
	url := "amqps://foo.bar:5671/"
	uri, err := ParseURI(url)
	if err != nil {
		t.Fatal("Could not parse")
	}

	if uri.String() != "amqps://foo.bar/" {
		t.Fatal("Defaults not encoded properly got:", uri.String())
	}
}

func TestURIDefaultPortAmqps(t *testing.T) {
	url := "amqps://foo.bar/"
	uri, err := ParseURI(url)
	if err != nil {
		t.Fatal("Could not parse")
	}

	if uri.Port != 5671 {
		t.Fatal("Default port not correct for amqps, got:", uri.Port)
	}
}
