## pr-slacker
Monitor an organizations pull requests, and provide notifications via Slack when a new PR is available for review.

#### Instructions

1. Download repository as ZIP archive and extract the files into a new folder.
2. Follow the [setup notes](#setup-notes) and [configuration instructions](#configuration-file).
3. Open command prompt and navigate to the folder you created (it must contain `config.json`).
4. Run the program in one of two ways:
    - Use the command `go run .\cmd\pr-slacker` to run the program without generating an executable file.
    - Use the command `go build .\cmd\pr-slacker` to compile the program. This will generate a new executable file `pr-slacker.exe`.   
        You can run this executable file anywhere so long as the `config.json` file is in the same directory.
        
#### Setup Notes
- You will need AWS credentials which are permitted to access DynamoDB.
- There should be a table on DynamoDB named `pull-requests` with a partition key labeled `pr_uid`.
- Your Slack bot will need to be added to your team workspace, with the necessary scope(s) to send messages.
- Your Slack bot will need to be added to the channel that it is configured to send messages in.

#### Configuration File

There is a file provided in the repository named `config.json.example`.

Before running the program, rename this file to `config.json` and open it in a text editor.

Fill in the file with your settings and save the file. After doing so, you can run the program successfully.

| Variable Name         | Type          | Description                                                                                                                                            |
| -------------         | ------------- | -------------                                                                                                                                          |
| github_username       | `string`      | Username of the Github account used to login and monitor data.                                                                                         |
| github_password       | `string`      | Password of the Github account used to login and monitor data.                                                                                         |
| github_organization   | `string`      | The Github account of the organization that you're monitoring pull requests from.                                                                      |
| github_save_cookies   | `bool`        | Whether or not your Github account session should be saved/restored in a local file automatically.                                                     |
| aws_access_key_id     | `string`      | AWS Access key used to authenticate your DynamoDB connection.                                                                                          |
| aws_access_key_secret | `string`      | AWS Secret key used to authenticate your DynamoDB connection.                                                                                          |
| aws_region            | `string`      | The [AWS region code](https://docs.aws.amazon.com/general/latest/gr/ddb.html#ddb_region) which is the host of your DynamoD database. (ex: `us-east-1`) |
| slack_oauth_token     | `string`      | OAuth token of your Slack application.                                                                                                                 |
| slack_channel_id      | `string`      | The ID of the Slack channel to post pull request notifications in.                                                                                     |
