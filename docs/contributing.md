# Contributing

First of all, thank you for donating your time improving this project. It is really appreciated.

## Setup your environment

To ensure the linter and tests are run before every commit, run the following command.
It will install a couple of tools used in the Makefile and setup the git pre-commit hooks.

You need to make sure your **GOPATH/bin** directory is included in your **PATH**.

```
make configure
```

## Makefile

This is a list of useful Makefile targets used for development.

| Target    | Description                      |
|-----------|----------------------------------|
| clean     | Clean the test reports directory |
| configure | Configure tools and git hooks    |
| godoc     | Start a local Godoc server       |
| lint      | Run the linter                   |
| qa        | Run the linter and tests         |
| test      | Run the tests                    |

## Coding style

The code follow the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).

## Pull requests

**Any pull requests that does not follow the house rules will not be accepted.**

### Linear history

The repository is configured to enforce a linear history. Rebase is your friend and you can read more 
about it [here](https://www.bitsnbites.eu/a-tidy-linear-git-history/#:~:text=A%20linear%20history%20is%20simply,branches%20with%20independent%20commit%20histories.).

### Commit comments

As the pull request will be squash merged into the **main** branch, please make sure 
the message is in the following format:

```
<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

The **type** must be one of:

| Type    | When to use                                                             | Outcome                                 |
|---------|-------------------------------------------------------------------------|-----------------------------------------|
| chore   | Clean the test reports directory                                        | No release                              |
| doc     | Clean the test reports directory                                        | No release                              |
| fix     | Configure tools and git hooks                                           | Release with **patch** version increase |
| feat    | Start a local Godoc server                                              | Release with **minor** version increase |
| perf    | *Never (due to the version tag action using this to trigger a release)* | Release with **major** version increase | 

**scope** is optional, but it should be equal to the name of the main package affected by the change.

**body** and **footer** are optional.

**subject** must: 
* be meaningful and consistent with the change being applied
* use the imperative mood

**footer** should contain any information about **breaking changes** and is also the place 
to reference GitHub issues that this commit reverts.
**Breaking changes** should start with the words *BREAKING CHANGE:* with a space or two newlines. 
The rest of the commit message is then used for this.

### Example of valid comments:

```
feat(breaker): add circuit breaker option
```

```
fix(breaker): resolve concurrency issue

<details>
```

```
feat(breaker): change signature of Do function

BREAKING CHANGE:


<explanation>
```
