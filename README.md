# pasmcli
pasmcli is a command line tool to manage, control and access secrets with Entrust Secrets Vault.

Entrust Secrets Vault provides centralized secure storage for managing and controlling access to secrets required to access systems and resources. Access to secrets are restricted to authorized applications and users.

You can securely store, manage, and access control secrets such as credentials, API keys, SSH keys, tokens, certificate private keys, and encryption keys. Secrets are managed, controlled and accessed through the pasmcli.

1. The pasmcli requires Entrust KeyControl version 5.2 or later.
2. The Secrets Vault must be created by the KeyControl Vault Administrator.
3. To manage the Secrets Vault using the pasmcli, you must be the admin for that Secrets Vault and have the login URL for the same.
4. All users authorized to access the Secrets Vault can use the pasmcli with the login URL.

## Releases

pasmcli's for Linux & Windows for each release can be found in Releases section (https://github.com/EntrustCorporation/pasmcli/releases)

## Build instructions

The code in this repo corresponds to the latest released version of pasmcli. In general, to use pasmcli, head over to Releases section to get pre-compiled binaries. If you do plan to build, follow instructions below.
1. Install go. pasmcli/Makefile expects to find go at /usr/local/go/bin/go
2. cd to pasmcli/
3. To build Linux & Windows cli binaries,
   
   ```$ gmake all```
5. To clean workspace,
   
   ```$ gmake clean```

For more information, see the Secrets Vault chapter in the Key Management Systems documentation at https://trustedcare.entrust.com/.
