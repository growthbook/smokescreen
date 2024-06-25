package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/stripe/smokescreen/cmd"
	"github.com/stripe/smokescreen/pkg/smokescreen"
)

// This default implementation of RoleFromRequest uses the CommonName of the
// client's certificate.  If no certificate is provided, the AllowMissingRole
// configuration option will control whether the request is rejected, or the
// default ACL is applied.
func defaultRoleFromRequest(req *http.Request) (string, error) {
	if req.TLS == nil {
		return "", smokescreen.MissingRoleError("defaultRoleFromRequest requires TLS")
	}
	if len(req.TLS.PeerCertificates) == 0 {
		return "", smokescreen.MissingRoleError("client did not provide certificate")
	}
	return req.TLS.PeerCertificates[0].Subject.CommonName, nil
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func startHealthCheckServer() {
	http.HandleFunc("/healthcheck", healthCheckHandler)
	// Run the health check server on a different port
	go func() {
		log.Fatal(http.ListenAndServe(":4751", nil))
	}()
}

func main() {
	// Register the health check route
	startHealthCheckServer()

	conf, err := cmd.NewConfiguration(nil, nil)
	if err != nil {
		logrus.Fatalf("Could not create configuration: %v", err)
	} else if conf != nil {
		conf.RoleFromRequest = defaultRoleFromRequest

		conf.Log.Formatter = &logrus.JSONFormatter{}

		adapter := &smokescreen.Log2LogrusWriter{
			Entry: conf.Log.WithField("stdlog", "1"),
		}

		// Set the standard logger to use our logger's writer as output.
		log.SetOutput(adapter)
		log.SetFlags(0)
		smokescreen.StartWithConfig(conf, nil)
	}
	// Otherwise, --help or --version was passed and handled by NewConfiguration, so do nothing
}
