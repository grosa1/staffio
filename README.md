# Staffio

An OAuth2 server with management for enterprise employees.


## features:

* All employees in LDAP.
* Login and general member settings.
* Reset password with email.
* Create, Edit and Remove employees with special manager.
* Client ID and Secret of all clients maintenance.
* Simplified content management for aritcles and links.
* A general OAuth2 authentication and authorization provider.
* Directly CAS implement for v1 and V2.

## APIs of oauth2

### Authorize (browse page)
> GET | POST /authorize

### Retrieve Token
> GET | POST /token

### Get Info
> GET | POST /info/{topic}


### APIs of <abbr title="Central Authentication Service">CAS</abbr>

| URI | Description |
| -------- | -------- |
| `/login` | credential requestor / acceptor |
| `/logout` | destroy CAS session (logout) |
| `/validate` | service ticket validation |
| `/serviceValidate` | service ticket validation [CAS 2.0] |
| `/proxyValidate` **TODO** | service/proxy ticket validation [CAS 2.0] |
| `/proxy` **TODO** | proxy ticket service [CAS 2.0] |
| `/p3/serviceValidate` **TODO** | service ticket validation [CAS 3.0] |
| `/p3/proxyValidate` **TODO** | service/proxy ticket validation [CAS 3.0] |


## Quick start

### Run all components as docker containers

````sh

# openldap
docker run --name staffio-ldap -p 389:389 -p 636:636 \
	-e LDAP_ADMIN_PASSWORD=mypassword \
	-d liut7/staffio-ldap:2.4.47-r2

# postgresql
docker create --name staffio-db-data -v /var/lib/postgresql busybox:1 echo staffio db data
docker run --name staffio-db -p 54322:5432 \
	-e DB_PASS=mypassword \
	-e TZ=Hongkong \
	--volumes-from=staffio-db-data \
	-d liut7/staffio-db:latest

# staffio main server
docker run --name staffio -p 3030:3030 \
	-e STAFFIO_BACKEND_DSN='postgres://staffio:mypassword@staffio-db/staffio?sslmode=disable' \
	-e STAFFIO_LDAP_HOSTS=slapd \
	-e STAFFIO_LDAP_BASE="dc=example,dc=org" \
	-e STAFFIO_LDAP_BIND_DN="cn=admin,dc=example,dc=org" \
	-e STAFFIO_LDAP_PASS='mypassword' \
	-e DEBUG='staffio:backends,staffio:ldap' \
	--link staffio-db --link staffio-ldap:slapd \
	-d liut7/staffio:latest web

# create a user as first staff and adminstrator
docker exec staffio staffio addstaff -u eagle -p mysecret -n eagleliut --sn liut
docker exec staffio staffio group -g keeper -a eagle

# now can open http://localhost:3030/ in browser

# add a oauth2 client (optional)
# demo client
echo "INSERT INTO oauth_client VALUES(1, '1234', 'Demo', 'aabbccdd', 'http://localhost:3000/appauth', '{}', now());" | docker exec -i staffio-db psql -Ustaffio staffio

````


## prepare development

### checkout

````sh
mkdir -p $GOPATH/src/github.com/liut
cd $GOPATH/src/github.com/liut
git clone https://github.com/liut/staffio.git
cd $GOPATH/src/liut/staffio
make dep
````

### environment

```
    cp -n .env.example .env
```

> `cat .env`
```
STAFFIO_HTTP_LISTEN=":3000"
STAFFIO_LDAP_HOSTS=slapd.hostname
STAFFIO_LDAP_BASE="dc=example,dc=org"
STAFFIO_LDAP_BIND_DN="cn=admin,dc=example,dc=org"
STAFFIO_LDAP_PASS="mypassword"
STAFFIO_BACKEND_DSN="postgres://staffio:mypassword@localhost:54322/staffio?sslmode=disable"
STAFFIO_PASSWORD_SECRET="mypasswordsecret"
```

## launch development

````sh
go get -u github.com/ddollar/forego
go get -u github.com/liut/rerun
npm install

forego start
````

## deployment

```sh
make dist package
scp dist/linux_amd64/staffio remote:/opt/staffio/bin/
rsync -rpt --delete templates htdocs remote:/opt/staffio/
```

### add staff
```sh
forego run ./staffio addstaff -u eagle -p mysecret -n eagle --sn eagle
```

## Plan

* <del>Peoples and groups sync with WxWork</del>
* <del>Signin with WxWork</del>
* Notification system
* Export for backup
* Batch import or restore from backup
