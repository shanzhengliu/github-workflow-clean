## Github Mass Workflow Clean Tool

### Introduction  
This tool is designed to clean up the workflow which has been disabled.  

### Prerequisites
1.  Install Docker in your machine.  
2.  Install Docker Compose in your machine.  
3.  Generate a personal access token which has permission to update workflow in your Github account.
    -   Click your profile icon in the top right corner of Github and select `Settings`.
    -   Go to `Developer settings` -> `Personal access tokens` ->  `Fine-grained tokens` -> `Generate new token`  
    -   Select `repo` and `workflow` in the `Select scopes` section.  
    -   Click `Generate token` and copy the token.

### How to use
1. Download the `docker-compose.yml` file.
2. Following the instruction, update the `GIHUB_XXX` environment variables in the `docker-compose.yml` file.
3. Run the following command to start the tool in the same directory as the `docker-compose.yml` file.
    ```shell
    docker-compose up
    ```

