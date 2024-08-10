# Logging into HashiCorp Cloud Platform and Setting Up Environment Variables

## Step 1: Sign Up or Log In to HashiCorp Cloud Platform

1. **Access HCP:**
   - Open your web browser and go to the [HashiCorp Cloud Platform](https://cloud.hashicorp.com/).

2. **Sign Up or Log In:**
   - If you already have an account, click on "Log In" and enter your credentials.
   - If you don't have an account, click on "Sign Up" and follow the prompts to create a new account.

## Step 2: Create a Service Principal

1. **Navigate to Service Principals:**
   - Once logged in, click on your organization name in the top-right corner, and then select "Service Principals" from the dropdown menu.

2. **Create a New Service Principal:**
   - Click on the "Create Service Principal" button.
   - Enter a name and description for the service principal (e.g., "Script Automation").
   - Click "Save".

3. **Generate Client ID and Secret:**
   - After creating the service principal, you'll be provided with a `Client ID` and `Client Secret`. Make sure to copy these values and store them securely. You will use these to authenticate API requests.

## Step 3: Retrieve Your Organization and Project IDs

1. **Open the HCP CLI:**
   - If you haven't already, install the HCP CLI by following the instructions [here](https://learn.hashicorp.com/tutorials/hcp/get-started-cli).

2. **Log in to the HCP CLI:**
   - Open your terminal and log in to the HCP CLI using your credentials:
     ```bash
     hcp login
     ```

3. **Retrieve Organization ID:**
   - Run the following command to display your HCP profile information:
     ```bash
     hcp profile display --format=json
     ```
   - Look for the `OrganizationID` in the output and copy it.

4. **Retrieve Project ID:**
   - In the same output, find and copy the `ProjectID`.

## Step 4: Set Up Environment Variables

You now need to set up environment variables in your terminal to allow your script to authenticate and interact with HCP Vault Secrets.

1. **Export Environment Variables:**
   - Open your terminal and export the following variables:
     ```bash
     export HCP_CLIENT_ID="your-client-id"
     export HCP_CLIENT_SECRET="your-client-secret"
     export HCP_ORG_ID="your-organization-id"
     export HCP_PROJECT_ID="your-project-id"
     export APP_NAME="your-app-name"
     ```

   - Replace `your-client-id`, `your-client-secret`, `your-organization-id`, `your-project-id`, and `your-app-name` with the actual values you obtained in the previous steps.

2. **Obtain HCP API Token:**
   - You need to get an API token to authenticate your requests. Run the following command:
     ```bash
     export HCP_API_TOKEN=$(curl https://auth.idp.hashicorp.com/oauth2/token \
         --data grant_type=client_credentials \
         --data client_id="$HCP_CLIENT_ID" \
         --data client_secret="$HCP_CLIENT_SECRET" \
         --data audience="https://api.hashicorp.cloud" | jq -r .access_token)
     ```

   - This command will set the `HCP_API_TOKEN` environment variable, which your script will use to authenticate with the HCP Vault Secrets API.

3. **Verify Environment Variables:**
   - You can verify that the environment variables are set correctly by running:
     ```bash
     echo $HCP_CLIENT_ID
     echo $HCP_CLIENT_SECRET
     echo $HCP_ORG_ID
     echo $HCP_PROJECT_ID
     echo $APP_NAME
     echo $HCP_API_TOKEN
     ```

   - Ensure that each command outputs the expected values.

## Step 5: Run Your Script

Now that the environment variables are set up, you can run your script:

```bash
ansiblesafe.go -o 5 --file=vault.yaml
```

This will execute your Go script using the environment variables you just set up, allowing it to authenticate with HCP and retrieve secrets as needed.
