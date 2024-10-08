# ansiblesafe
ansiblesafe is a simple Go script that makes it easy to encrypt and decrypt YAML files using the Ansible Vault CLI. With a user-friendly menu option, you can easily secure your secrets in an Ansible playbook. This has been custom build to store Red Hat Credentials to be used with Ansible playbooks. 

![Build Status](https://github.com/tosin2013/ansiblesafe/actions/workflows/build.yml/badge.svg)
![Release Status](https://github.com/tosin2013/ansiblesafe/actions/workflows/release.yml/badge.svg)
[![Test AnsibleSafe](https://github.com/tosin2013/ansiblesafe/actions/workflows/test.yml/badge.svg)](https://github.com/tosin2013/ansiblesafe/actions/workflows/test.yml)

![20230211185054](https://i.imgur.com/gsItHDF.png)


## Features
* Generate ansible vault files to be used for playbooks 
* Pull secrets from HashiCorp Vault and create ansible vault files
* write ansible vault values to HashiCorp Vault
  
## Outputs

### Encrypted Result 
```
$ cat ~/vault.yml 
$ANSIBLE_VAULT;1.1;AES256
36636639646139323163653635303639646266313532623937333264353464383434386432643331
3930653138613130633864313363626236356136356266330a646336656438653937333434306638
34373033656162626433633231366563393565646235663439623037363235363831666433623266
6530613664343764650a306664663135306265616434313733366261313438323139613964613433
35336665346232383831626132633137316136336337663364393065616636663063306536346337
35313539383166326266346135393265306535383062643931333831333238396363613563373735
64653537383964313933663166386137616532643233303566343330333563336430356161363665
30643731323963303730316466356438636363343230366261666263396431313162373961313866
37613438386431323137643666303634356135396235653861626434356437383461643661643662
66383636646339666232653263303762623066386634306565336133663266306335663364383733
65303137313061636336346664383138313962356533633038623830316264666539653933386161
36396637356336636265323437613037386639386564343039323662393461343634623864336666
30363265613535313631383538663764623864613839366134623164313733333132646139616637
3831316663636239653234653430383633666234383036653361
```

### Decrypted Result 
```
$ cat ~/vault.yml 
rhsm_username: rheluser
rhsm_password: rhelpassword
rhsm_org: orgid
rhsm_activationkey: activationkey
admin_user_password: password
offline_token: offlinetoken
automation_hub_offline_token: automationhubtoken
openshift_pull_secret: pullsecret
freeipa_server_admin_password: password
aws_access_key: accesskey
aws_secret_key: secretkey
```

## Quick Start 
```
dnf install ansible-core -y 
curl -OL https://github.com/tosin2013/ansiblesafe/releases/download/v0.0.8/ansiblesafe-v0.0.8-linux-amd64.tar.gz
tar -zxvf ansiblesafe-v0.0.8-linux-amd64.tar.gz
chmod +x ansiblesafe-linux-amd64 
sudo mv ansiblesafe-linux-amd64 /usr/local/bin/ansiblesafe
```

### Handling AWS Keys
To use AWS keys (`aws_access_key` and `aws_secret_key`) in your Ansible playbooks, ensure they are included in your `vault.yml` file as shown in the "Decrypted Result" section. These keys are crucial for interacting with AWS services and should be handled with the same security measures as other sensitive information.

## Menu Options 
**If you do not pass any flags everything wil be auto generated for you**
```
$ ansiblesafe -h
Usage of /tmp/go-build1657505477/b001/exe/ansiblesafe:
  -f, --file string     Path to YAML file (default: $HOME/vault.yml)
  -o, --operation int   Operation to perform (1: encrypt, 2: decrypt, 3: Write secrets to HashiCorp Vault, 4: Read secrets from HashiCorp Vault, 5: skip encrypting/decrypting)
```

## Hasicorp Examples 
**Write secrets to HashiCorp Vault**
```
$ export VAULT_ADDRESS=http://127.0.0.1:8200/
$ export VAULT_TOKEN=token
$ export SECRET_PATH=ansiblesafe/example
$ ansiblesafe -o 3
```

**Read secrets from HashiCorp Vault and safe to vault.yaml**
```
$ export VAULT_ADDRESS=http://127.0.0.1:8200/
$ export VAULT_TOKEN=token
$ export SECRET_PATH=ansiblesafe/example
$ ansiblesafe -o 4
$ ansiblesafe -o 1 # Optional encrypt the file
```

## Usage
**Instructions to use ansiblesale without a password prompt**
```
$ touch ~/.vault_password
$ chmod 600 ~/.vault_password
# The leading space here is necessary to keep the command out of the command history
$  echo password >> ~/.vault_password
# Link the password file into the current working directory
$ ln ~/.vault_password .
# Set the environment variable to the location of the file
$ export ANSIBLE_VAULT_PASSWORD_FILE=.vault_password
```

## Hashicorp Cloud Instructions
* [Logging into HashiCorp Cloud Platform and Setting Up Environment Variables](docs/hashicorp_cloud_setup.md)
* [Setting Up Variables in HashiCorp Cloud Platform (HCP) Vault Secrets](docs/hashicorp_cloud_secret_setup.md)
* [HashiCorp Cloud](https://portal.cloud.hashicorp.com/)
```
$ export HCP_CLIENT_ID="your-client-id"
$ export HCP_CLIENT_SECRET="your-client-secret"
$ export HCP_ORG_ID=$(hcp profile display --format=json | jq -r .OrganizationID)
$ export HCP_PROJECT_ID=$(hcp profile display --format=json | jq -r .ProjectID)
$ export APP_NAME="your-app-name"
$ ansiblesafe -o 5  --file=vault.yaml
# Encrypt the file
$ ansiblesafe -o 1 --file=vault.yaml
# Decrypt the file
$ ansiblesafe -o 2 --file=vault.yaml
```


## Requirements
* Ansible Vault CLI

## Deveploer requirements - WIP
* [Go](https://gist.github.com/tosin2013/d4f4420231a96aed2116efb4d6b151a0)
* git
* ansible-core
```
git clone https://github.com/tosin2013/ansiblesafe.git
cd ansiblesafe
```
### run app
``` 
go run ansiblesafe.go
```

## Documentation 
[Ansible Vault Secrets Documentation](https://dev.to/tosin2013/ansible-vault-secrets-documentation-3g1a)

## Contributing
This project is open source and contributions are welcome! If you have any suggestions or bug reports, please open an issue or create a pull request.

## License


## Authors
* Tosin Akinosho - [tosin2013](https://github.com/tosin2013)


