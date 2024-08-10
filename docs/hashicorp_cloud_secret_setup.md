# Setting Up Variables in HashiCorp Cloud Platform (HCP) Vault Secrets

This guide will walk you through the process of setting up and storing the following variables in HashiCorp Cloud Platform (HCP) Vault Secrets:

- `rhsm_username`: rheluser
- `rhsm_password`: rhelpassword
- `rhsm_org`: orgid
- `rhsm_activationkey`: activationkey
- `admin_user_password`: password
- `offline_token`: offlinetoken
- `automation_hub_offline_token`: automationhubtoken
- `openshift_pull_secret`: pullsecret
- `freeipa_server_admin_password`: password
- `aws_access_key`: accesskey
- `aws_secret_key`: secretkey

## Prerequisites

Before you begin, ensure that you have the following:

1. **HCP Account:** Sign up or log in to your [HashiCorp Cloud Platform account](https://cloud.hashicorp.com/).
2. **HCP CLI:** Installed HCP CLI on your local machine.
   - Installation instructions can be found [here](https://learn.hashicorp.com/tutorials/hcp/get-started-cli).
3. **Service Principal:** A service principal created in HCP, with the `Client ID` and `Client Secret` stored securely.
4. **jq:** Installed `jq` for parsing JSON in shell scripts.

## Step 1: Create a New Application in HCP

1. **Log in to HCP:**
   - Go to [HashiCorp Cloud Platform](https://cloud.hashicorp.com/) and log in.

2. **Navigate to the Vault Secrets:**
   - Select the "Vault" option from the HCP dashboard.

3. **Create a New Application:**
   - Within the Vault Secrets interface, create a new application by clicking on "Create App."
   - Provide a name for your application, such as `MyAppSecrets`.
   - Click "Create" to finalize the application.

## Step 2: Set Up Environment Variables

Set the following environment variables on your local machine to interact with HCP and store secrets.

```bash
export HCP_CLIENT_ID="your-client-id"
export HCP_CLIENT_SECRET="your-client-secret"
export HCP_ORG_ID=$(hcp profile display --format=json | jq -r .OrganizationID)
export HCP_PROJECT_ID=$(hcp profile display --format=json | jq -r .ProjectID)
export APP_NAME="MyAppSecrets"
```

## Step 3: Authenticate and Obtain an API Token

Use your `Client ID` and `Client Secret` to obtain an API token from HCP:

```bash
export HCP_API_TOKEN=$(curl https://auth.idp.hashicorp.com/oauth2/token \
     --data grant_type=client_credentials \
     --data client_id="$HCP_CLIENT_ID" \
     --data client_secret="$HCP_CLIENT_SECRET" \
     --data audience="https://api.hashicorp.cloud" | jq -r .access_token)
```

## Step 4: Store Variables in HCP Vault Secrets

You can now store the required variables in HCP Vault Secrets. Use the following commands to store each variable:

```bash
curl \
    --location "https://api.cloud.hashicorp.com/secrets/2023-06-13/organizations/$HCP_ORG_ID/projects/$HCP_PROJECT_ID/apps/$APP_NAME/secrets" \
    --request POST \
    --header "Authorization: Bearer $HCP_API_TOKEN" \
    --header "Content-Type: application/json" \
    --data-raw '{
        "secrets": [
            {
                "name": "rhsm_username",
                "value": "rheluser"
            },
            {
                "name": "rhsm_password",
                "value": "rhelpassword"
            },
            {
                "name": "rhsm_org",
                "value": "orgid"
            },
            {
                "name": "rhsm_activationkey",
                "value": "activationkey"
            },
            {
                "name": "admin_user_password",
                "value": "password"
            },
            {
                "name": "offline_token",
                "value": "offlinetoken"
            },
            {
                "name": "automation_hub_offline_token",
                "value": "automationhubtoken"
            },
            {
                "name": "openshift_pull_secret",
                "value": "pullsecret"
            },
            {
                "name": "freeipa_server_admin_password",
                "value": "password"
            },
            {
                "name": "aws_access_key",
                "value": "accesskey"
            },
            {
                "name": "aws_secret_key",
                "value": "secretkey"
            }
        ]
    }'
```

This command creates and stores each secret under the specified names in your HCP Vault Secrets application.

## Step 5: Verify Stored Secrets

To verify that your secrets have been successfully stored, you can list them using the following command:

```bash
curl \
    --location "https://api.cloud.hashicorp.com/secrets/2023-06-13/organizations/$HCP_ORG_ID/projects/$HCP_PROJECT_ID/apps/$APP_NAME/open" \
    --request GET \
    --header "Authorization: Bearer $HCP_API_TOKEN" | jq
```

This command will display the stored secrets in your application.

## Conclusion

You have now successfully set up and stored the necessary variables in HashiCorp Cloud Platform (HCP) Vault Secrets. These secrets are securely stored and can be accessed by your applications as needed.

For more advanced usage and integration with your applications, refer to the official [HashiCorp Cloud Platform documentation](https://learn.hashicorp.com/cloud).
