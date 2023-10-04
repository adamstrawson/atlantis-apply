# Atlantis apply en masse 

This utility allows you to mass approve and apply [Renovate](https://www.mend.io/renovate/) PRs within [Atlantis](https://www.runatlantis.io/)

## Prerequisites

Before using this utility, make sure you have the following:

- Go installed on your system: [Download and Install Go](https://golang.org/dl/)
- A GitHub personal access token: [Creating a personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)

## Compilation

1. Clone this repository:

   ```bash
   git clone https://github.com/adamstrawson/atlantis-apply.git
   ```

2. Change to the project directory:

    ```bash
    cd atlantis-apply
    ```

3. Build the Go program:

    ```bash
    go build
    ```


## Usage

You can provide the repository name and token using either command-line flags or environment variables.

### Using Command-Line Flags
    
```bash
./atlantis-apply -repo=adamstrawson/terraform -token=foobar
```

### Using Environment Variables
You can also set the following environment variables:

```bash
export GITHUB_REPO=adamstrawson/terraform
export GITHUB_TOKEN=foobar
```

Once the environment variables are set, run the program without flags:

```bash
./atlantis-apply
```
