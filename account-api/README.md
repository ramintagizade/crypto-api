
## Create Docker image of the golang application

<code> docker build -t account-api .  </code>

## Run the container 

<code> docker run -p 9123:9123 -it account-api </code>
 



## Start PostgreSQL from Docker 

<code> docker-compose up -d </code>

### Run inside the container
<code>docker-compose run db bash </code>

### Connect to database  
<code>psql  --host=db --username=admin --dbname=crypto</code>

### Connect to database outside 
<code> psql --host=localhost --username=admin --dbname=crypto</code>