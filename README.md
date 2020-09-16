# rdr-emailer

Generic tool for sending emails to DR system users via SMTP.

## Requirement

The [GO](https://golang.org) compiler is required to build the source code.

## Build

```bash
$ make
```

The executable `rdr-emailer.linux_amd64` will be created right in the current directory.

## Release on GitHub 

```bash
$ VERSION=<version> make github-release 
```

It will create a new release with the `<version>` number and add the executable `rdr-emailer.linux_amd64` as an asset of the release.

It requires authentication to Github using the Github personal token.

## Usage

Firstly define an email template, an example can be found in the [template.txt](template.txt) file.  Secondly make a recipient list as shown in the example file [recipients.csv](recipients.csv).

An example script [scripts/rdr-get-email-recipients.sh](scripts/rdr-get-email-recipients.sh) can be used to dump the `recipients.csv` file containing all repository users.  The script makes use of the `iquest` command.

Run the following command to send the email to all recipients:

```bash
$ ./rdr-emailer -f from_address -l recipients.csv -n smtp_host -p smtp_port -u smtp_username -s smtp_password template.txt
```

The command-line options are provided with `-h` option:

```bash

Usage: ./rdr-emailer [OPTIONS] <template>

OPTIONS:
  -f email
    	set the sender's email address. (default "no-reply@donders.ru.nl")
  -l path
    	set path of the file containing a list of recipients.
  -n hostname
    	set the network hostname of the SMTP server. (default "smtp-auth.ru.nl")
  -p port
    	set the network port of the SMTP server. (default 587)
  -s password
    	set SMTP password for PLAIN authentication.
  -u username
    	set SMTP username for PLAIN authentication.
```

## Docker container

The [Dockerfile](Dockerfile) is also provided to run rdr-emailer using a docker container.

To build the container, run

```bash
$ docker build -t rdr-emailer --force-rm .
```

Simply run the container a help message will be shown.

```bash
$ docker run rdr-emailer
```

The following example command shows how to run the rdr-emailer via the container.

```bash
$ docker run -v `pwd`/recipients.csv:/recipients.csv -v `pwd`/template.txt:/template.txt rdr-emailer /rdr-emailer -f from_address -l /recipients.csv -n smtp_host -p smtp_port -u smtp_username -s smtp_password /template.txt
```

Note that in the command above, we bind-mount `recipients.csv` and `template.txt` in the present working directory (i.e. `pwd`) into `/recipients.csv` and `/template.txt` respectively in the container.
