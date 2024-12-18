# HTTP router (museum/http)

The custom HTTP Router in mūsēum is designed to handle complex routing requirements, including specific path matching and proxying requests to containers. It supports flexible path matching, allowing for both exact matches and forwarding of "rest-path" segments to proxied services. This custom solution was necessary due to the lack of support for such functionality in existing HTTP libraries.  

The router leverages the Go standard library for the server implementation, focusing on the routing logic. It ensures that API endpoints, health probes, and service proxying are handled efficiently and correctly, providing a robust and flexible routing mechanism tailored to the needs of the mūsēum project.