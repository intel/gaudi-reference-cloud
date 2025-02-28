rm *.crt
rm *.csr
rm *.key

# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -sha256 -newkey rsa:4096 -keyout ca.key -out ca.crt -days 356 -nodes -subj "/C=IL/CN=*.intel.com"

echo "CA's self-signed certificate"
openssl x509 -in ca.crt -noout -text

# 2. Generate web server's private key and certificate signing request (CSR)
openssl req -new -newkey rsa:4096 -keyout server.key -out server.csr -nodes -subj '/CN=*.intel.com'

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -sha256 -days 365 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt -extfile server-ext.cnf

echo "Server's signed certificate"
openssl x509 -in server.crt -noout -text

# 4. Generate client's private key and certificate signing request (CSR)
openssl req -new -newkey rsa:4096 -keyout client.key -out client.csr -nodes -subj '/CN=MaasGateway'

# 5. Use CA's private key to sign client's CSR and get back the signed certificate
openssl x509 -req -sha256 -days 365 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 02 -out client.crt # -extfile client-ext.cnf

echo "Client's signed certificate"
openssl x509 -in client.crt -noout -text


#rm *.pem
#
## 1. Generate CA's private key and self-signed certificate
#openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=IL/CN=*.intel.com"
#
#echo "CA's self-signed certificate"
#openssl x509 -in ca-cert.pem -noout -text
#
## 2. Generate web server's private key and certificate signing request (CSR)
#openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=IL/CN=*.intel.com"
#
## 3. Use CA's private key to sign web server's CSR and get back the signed certificate
#openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf
#
#echo "Server's signed certificate"
#openssl x509 -in server-cert.pem -noout -text
#
## 4. Generate client's private key and certificate signing request (CSR)
#openssl req -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem -subj "/C=IL/O=MaasGateway/CN=*.intel.com"
#
## 5. Use CA's private key to sign client's CSR and get back the signed certificate
#openssl x509 -req -in client-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile client-ext.cnf
#
#echo "Client's signed certificate"
#openssl x509 -in client-cert.pem -noout -text
