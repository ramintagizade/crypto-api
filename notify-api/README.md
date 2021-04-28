
## Create Docker image of the golang application

<code> docker build -t notify-api .  </code>

## Run the container 

<code> docker run -p 8080:8081 -it notify-api </code>


## Start RabbitMQ server from docker 

<code> docker-compose up -d </code>
