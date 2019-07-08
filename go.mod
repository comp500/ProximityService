module github.com/comp500/proximityservice

go 1.12

require (
	github.com/gobuffalo/packr/v2 v2.5.2
	github.com/gorilla/websocket v1.4.0
	github.com/paypal/gatt v0.0.0-20151011220935-4ae819d591cf
)

// Depending on a PR that has a fix for a bug in packr2
replace github.com/gobuffalo/packr/v2 => github.com/nlepage/packr/v2 v2.2.1-0.20190515165437-bf153a1f0078
