# Admin Api/UI

The Admin API serves data to the Admin UI.

## Running the Server

#### Env Var
`ADMIN_API_JWT_SECRET` is required.

#### Running

```bash
cd /workspaces/admin-api
go run . -p8081
```

## Running the UI

#### Env file
Create `.env` file in the home directory with the properly specified `REACT_APP_API_DOMAIN`. This is the URL to the `admin-api`. Please check `.env.example.`

#### Running
```bash
cd /workspaces/admin-ui
npm run start
```

## Development

#### Ports

In Codespaces, after the `admin-api` is running, you need to make the Port Visibility public to be accessible from the `admin-ui`. Go to the `Port` tab, right-click on the `admin-api` port (e.g., 8081), and then choose `Port Visibility` -> `Public`.

#### User permissions & first time login

* Run the migrations
```bash
cd /workspaces/server
go run . migrate
```

* Once the UI and the API are running, the only thing required to be able to log in is the Metamask plugin installed. When you click on `Login With Metamask`, a new user and orgnisation with your wallet will be created (if you are logging in for the first time). To have root permissions, you will need to update your permission group manually in the database: 

`UPDATE users SET user_group_id=1 WHERE wallet_address='0xd694cfc8c66e34371eae8ebe03d54867e5c6cec4'` 

Note: use your address
