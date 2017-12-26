# How to contribute

First of all, thank you for showing interest in this project! Collaboration
is an important part in any project, and your contribution is important
regardless of how big or small it is. :raised_hands:

This project aims to provide 3 things:

- Code with high quality
- Documentation that is easy to digest
- Smooth collaboration

To be able to provide these 3 things, a set of guidelines needs to exist that all
contributors adhere to:

- We write tests for our code
- We document our code
- Everyone is welcome to contribute
- There are no stupid questions

## Versioning

The version of the project is kept in the file **VERSION** in the root of the
project. The version consists of 3 parts: **major**, **minor** and **patch**.
Written as `{major}.{minor}.{patch}`.

The version **only** tracks changes done to the functionality of the library.
Changes that does not alter the functionality of the library
(such as documentation improvement, GitHub templates, etc) does not alter
the version.

Changes that fix bugs or improve current features alters the **patch** part of
the version number.

Changes that add new features, alters the **minor** part of the version number.

Changes to the library API in a non-compatible way alters the **major** part of
the version number. Usually this means a complete rewrite of the library.

Until the library reaches version `1.0.0`, the library API will change in
non-compatible ways.

## Submitting changes

This project uses the *master*-branch as the latest stable branch and
contributors should strive to get their changes merged into the *master*-branch
as soon as possible.

For changes to be allowed to be merged into the *master*-branch they need to:

- Bump the version in the file **VERSION**
- Document the new version in **CHANGELOG.md**
- Be submitted as a GitHub pull-request
- Pass the automated tests on Travis CI
- Be reviewed by another project member

A pull-request can have multiple commits, but each commit should only make
**one** kind of change, such as:

- Fixing a bug
- Adding a feature

By doing this we can more easily pinpoint where a problem is introduced into the
code base.

A GitHub pull-request template is used, so each time you make a PR, you'll get
a checklist with the requirements.
