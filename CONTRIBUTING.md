# Contributing to go-version

If you are looking to contribute to this project and want to open a GitHub pull request ("PR"), there are a few guidelines of what we are looking for in patches. Make sure you go through this document and ensure that your code proposal is aligned.

## Sign off your work

The `sign-off` is an added line at the end of the explanation for the commit, certifying that you wrote it or otherwise have the right to submit it as an open-source patch. By submitting a contribution, you agree to be bound by the terms of the DCO Version 1.1 and Mozilla Public License Version 2.0.

Signing off a commit certifies the below Developer's Certificate of Origin (DCO):

```text
Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

   (a) The contribution was created in whole or in part by me and I
       have the right to submit it under the open source license
       indicated in the file; or

   (b) The contribution is based upon previous work that, to the best
       of my knowledge, is covered under an appropriate open source
       license and I have the right under that license to submit that
       work with modifications, whether created in whole or in part
       by me, under the same open source license (unless I am
       permitted to submit under a different license), as indicated
       in the file; or

   (c) The contribution was provided directly to me by some other
       person who certified (a), (b) or (c) and I have not modified
       it.

   (d) I understand and agree that this project and the contribution
       are public and that a record of the contribution (including all
       personal information I submit with it, including my sign-off) is
       maintained indefinitely and may be redistributed consistent with
       this project or the open source license(s) involved.
```

All contributions to this project are licensed under the [Mozilla Public License, version 2.0](https://www.mozilla.org/en-US/MPL/2.0/).

When committing your change, you can add the required line manually so that it looks like this:

```text
Signed-off-by: John Doe <john.doe@example.com>
```

Alternatively, configure your Git client with your name and email to use the `-s` flag when creating a commit:

```text
$ git config --global user.name "John Doe"
$ git config --global user.email "john.doe@example.com"
```

Creating a signed-off commit is then possible with `-s` or `--signoff`:

```text
$ git commit -s -m "this is a commit message"
```

To double-check that the commit was signed-off, look at the log output:

```text
$ git log -1
commit 37ceh170e4hb283bb73d958f2036ee5k07e7fde7 (HEAD -> issue-35, origin/main, main)
Author: John Doe <john.doe@example.com>
Date:   Mon Aug 1 11:27:13 2020 -0400

    this is a commit message

    Signed-off-by: John Doe <john.doe@example.com>
```


[//]: # (TODO: Commit guidelines, granular commits)


[//]: # (TODO: Commit guidelines, descriptive messages)


[//]: # (TODO: Commit guidelines, commit title, extra body description)


[//]: # (TODO: PR title and description)

## Test your changes

This project has a `Makefile` with helpers for running tests and lint checks. Although PRs will have
automatic checks for these, it is useful to run them locally, ensuring they pass before submitting
changes. Any tooling a task needs is downloaded into `.tool/` on first use, so there is nothing to
install up front:

```text
$ make unit
$ make static-analysis
```

Run `make help` to see every available task.

## Upstream changes

This repository is a fork of [hashicorp/go-version](https://github.com/hashicorp/go-version).
If your change is not specific to Anchore, please open it upstream first so that it
reaches everyone; this fork picks upstream changes up on the next sync.

When syncing upstream, Anchore's divergences take precedence and must be preserved:

- `Constraints` is a 2D slice supporting `||` (or) constraints
- the `^` (caret) and `~` (tilde) constraint operators
- `MarshalJSON`/`UnmarshalJSON` on `Version` and `Constraints`
- comparison operators do not apply `prereleaseCheck`, per
  [semver spec item 11](https://semver.org/#spec-item-11)

A sync is a real merge of `hashicorp/go-version` into `main`, and must be merged with a merge
commit rather than squashed or rebased. Squashing collapses the merge into a single-parent commit,
which drops the upstream authorship the MPL-2.0 requires us to preserve and, more practically,
loses the merge base: `git merge-base` would fall back to the original 2020 fork point and every
later sync would replay the same conflicts.

## Document your changes

When proposed changes are modifying user-facing functionality or output, it is expected the PR will include updates to the documentation as well.
