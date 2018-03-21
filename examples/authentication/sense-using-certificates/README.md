# Authentication: Qlik Sense certificates and headers

This example will show you how you use the Qlik Sense Enterprise root certificates
and HTTP headers to impersonate a user. This is useful when creating services
or scripts that needs to run as different users against your deployment.

Please note that this approach uses root certificates that can be used to impersonate
_anyone_ in your Qlik Sense Enterprise deployment, be careful. It also bypasses the
Qlik Sense Proxy authentication/load-balancing.

## Prerequisites

To run this example you need to export certificates from your Sense Enterprise
deployment, this is possible to do either via API or through the QMC (PEM format).

Once you have the certificates, place them in the ./certificates folder and modify
the runnable code with the appropriate parameters (highlighted using comments in the
example code).

## Runnable code

* [Sense using certificates](./sense-using-certificates.go)

## Documentation

* [Qlik Sense Help: Exporting certificates](http://help.qlik.com/en-US/sense/June2017/Subsystems/ManagementConsole/Content/export-certificates.htm)
* [Qlik Sense Help: Certificate architecture](http://help.qlik.com/en-US/sense/June2017/Subsystems/PlanningQlikSenseDeployments/Content/Deployment/Server-Security-Authentication-Certificate-Trust-Architecture.htm)
