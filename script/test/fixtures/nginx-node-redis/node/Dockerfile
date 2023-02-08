FROM centos:7

RUN yum -y install epel-release && yum install -y nodejs npm gcc* make 
RUN /bin/bash -c 'npm install -g nodemon' && mkdir /src

# Define working directory
WORKDIR /src
ADD . /src

RUN cd /src && npm install

# Expose port
EXPOSE  8080

# Run app using nodemon
CMD /bin/bash -c 'nodemon /src/index.js'
