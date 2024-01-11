package templates

import (
	"encoding/yaml"
	"strings"
	"uuid"

	corev1 "k8s.io/api/core/v1"
)

#ConfigMap: corev1.#ConfigMap & {
	#config:    #Config
	apiVersion: "v1"
	kind:       "ConfigMap"
	metadata: {
		name:      "\(#config.metadata.name)-\(_checksum)"
		namespace: #config.metadata.namespace
		labels:    #config.metadata.labels
		if #config.metadata.annotations != _|_ {
			annotations: #config.metadata.annotations
		}
	}
	immutable: true
	let _checksum = strings.Split(uuid.SHA1(uuid.ns.DNS, yaml.Marshal(data)), "-")[0]
	data: {
		"nginx.default.conf": """
			server {
				listen       8080;
				server_name  \(#config.metadata.name);

				location / {
			  	root   /usr/share/nginx/html;
			  	index  index.html index.htm;
				}

				location /healthz {
					access_log off;
					default_type text/plain;
					return 200 "OK";
				}

				error_page  404              /404.html;
			  error_page  500 502 503 504  /50x.html;
			  location = /50x.html {
			    root   /usr/share/nginx/html;
			  }
			}
			"""
		"index.html":         """
			<!doctype html>
			<html lang="en">
			<head>
			 	<meta charset="utf-8">
			 	<meta http-equiv="refresh" content="10" />
				<title>\(#config.metadata.name)</title>
				<style>
				html { color-scheme: light dark; }
				body { width: 35em; margin: 0 auto;
				font-family: Tahoma, Verdana, Arial, sans-serif; }
				</style>
			</head>
			<body>
				<h1> \(#config.message) from \(#config.metadata.name)!</h1>
				<p>If you see this page, the <b>\(#config.metadata.name)</b> instance is successfully deployed in the <b>\(#config.metadata.namespace)</b> namespace by Timoni.</p>
			</body>
			</html>
			"""
	}
}
