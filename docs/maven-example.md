# Fabric8 Maven Plugin + Kompose:
Let's deploy a Springboot Java application with Docker Compose file using Fabric8 Maven Plugin to Kubernetes or OpenShift.

##### Requirements
* Linux or MacOS or Windows
* JDK 1.7+ - [JDK Quick Installation Guide](http://openjdk.java.net/install/)
* Maven 3.x+ - [Maven Installation Guide](http://www.baeldung.com/install-maven-on-windows-linux-mac) 
* Kompose - [Kompose Installation Guide](/docs/installation.md)

__1. Clone the example project from GitHub__
```bash
$ git clone https://github.com/piyush1594/kompose-maven-example.git
```

Change current directory to `kompose-maven-example` directory.
```bash
$ cd kompose-maven-example
```

__2. Add Fabric8 Maven Plugin to your project__     
```bash
$ mvn io.fabric8:fabric8-maven-plugin:3.5.28:setup
```

Add the Fabric8 Maven Plugin configuration to `pom.xml` of project. `pom.xml` is manifest or deployment descriptor file of a maven project.

__3. Install Kompose through Maven__
```bash
$ mvn fabric8:install
```

This command installs the `kompose` on the host.

__4. Configure Fabric8 Maven Plugin to use a Docker Compose file__
```bash
<plugin>
  <groupId>io.fabric8</groupId>
  <artifactId>fabric8-maven-plugin</artifactId>
  <configuration>
    <composeFile>path for docker compose file</composeFile>
  </configuration>
  <executions>
    <execution>
      <goals>
        <goal>resource</goal>
        <goal>build</goal>
      </goals>
    </execution>
  </executions>
</plugin>
```

Add the `<configuration>` and `<executions>` sections to `pom.xml` as shown in above `pom.xml` snippet. Update the `<composeFile>` to provide the relative path of Docker Compose file from `pom.xml`

__5. Deploy application on Kubernetes or OpenShift__

 Make sure that Kubernetes/OpenShift cluster or Minikube/minishift is running. In case, if anything of this is not running, you can run minishift to test this application by using following command. 
```bash
$ minishift start
```

 Below command deploys this application on Kubernetes or OpenShift.
```bash
$ mvn fabric8:deploy  
```

Now that your service has been deployed, let's access it by querying `pod`, `service` from Kubernetes or OpenShift.
```bash
$ oc get pods
NAME                                    READY     STATUS      RESTARTS   AGE
springboot-docker-compose-1-xl0vb       1/1       Running     0          5m
springboot-docker-compose-s2i-1-build   0/1       Completed   0          7m
```

```bash
$ oc get svc
NAME                        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
springboot-docker-compose   172.30.205.137   <none>        8080/TCP   6m
```

Let's access the Springboot service.
```bash
$ minishift openshift service --in-browser springboot-docker-compose
Created the new window in existing browser session.
```

It will open your application endpoint in default browser.

![Output-Diagram](/docs/images/kompose-maven-output-diagram.png)
