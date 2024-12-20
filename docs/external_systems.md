# External systems

mūsēum is designed to not be hard coupled to any external systems. This means that in order to connect to another application, we utilize an [event driven](https://learn.microsoft.com/en-us/azure/architecture/guide/architecture-styles/event-driven) approach. The events are handled by NATS which is used as a simple pub/sub service.

Events are emitted whenever an exhibit is created, deleted, started or stopped. The events themselves use the [cloudevents spec](https://cloudevents.io). One example for such an external system is the [phaidra-connect](https://github.com/phaidra/museum-phaidra-connect) service used to import exhibits into Phaidra.