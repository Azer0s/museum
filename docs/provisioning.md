# Provisioning

Provisioning is done in 4 basic steps. Persistence in mūsēum will always lock pessimistically to ensure distributed data consistency. The steps are as followed:

1. **Look up the exhibit in etcd**

2. **Check if the exhibit is running**
3. **Start the exhibit**
   1.  **Starting step**
      1. **Validity check** - the environment is checked for its validity. Volumes to mount, networks, etc. are checked.
      2. **Cleanup** - If a network or container exists in a shutdown state, they will be cleaned up here.
      3. **Start containers** - Containers are started sequentially.
         1. **Livechecks** - If the container has a livecheck defined, it will be run here. If the livecheck fails or takes too long, the startup process will be aborted.
   2. **Running step** - If the application started correctly, its state will be set to `running`.
4. **Renew lease** - Whenever the application is accessed, the lease will be renewed keeping the application from being cleaned up.

