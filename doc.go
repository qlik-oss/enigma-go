// Package enigma is a library that helps you communicate with Qlik Associative Engine.
// Examples of use may be building your own analytics tools, back-end services, or other tools communicating with Qlik Associative Engine.
//
// # Schemas
//
// enigma-go includes generated API code that is based on the latest available Qlik Associative Engine schema.
// When a new schema is available a new version of enigma-go will be made available
//
// # Getting started
//
// Connecting to Qlik Associative Engine and intreract with a document/app involves at least the following steps:
//
// 1. Create and setup a Dialer object with TLS configuration etc
//
// 2. Open the WebSocket to Qlik Associative Engine using the Dial function in the Dialer
//
// 3. Open or create a document/app using openDoc or createApp
//
// See the example below for an illustration of how it may look. For more detail examples look at
// the examples in https://github.com/qlik-oss/enigma-go/tree/master/examples. See respective README.md file for further information
package enigma
