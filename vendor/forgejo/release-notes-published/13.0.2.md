Forgejo v13.0.2 contains critical security fixes. Originally scheduled for 7 November, the release date of these patches was advanced because a vulnerability had been leaked publicly.

### Vulnerability (**Critical**): prevent writing to out-of-repo symlink destinations while evaluating template repos
- [CVSS 9.5 Critical](https://www.first.org/cvss/calculator/4-0#CVSS:4.0/AV:N/AC:H/AT:P/PR:N/UI:N/VC:H/VI:H/VA:H/SC:H/SI:H/SA:H)
- [Patch](https://codeberg.org/forgejo/forgejo/commit/449b5bf10e3ba4baf2ef33b535b5e8ba15711838)

When creating a repository based upon a template repository, Forgejo reads the contents of the template repository files, expands variables within the file content, and writes a new file for use in the newly created repository. In the event that the template repository file was a symlink to a file outside of the repository, Forgejo would follow this symlink to read, expand content, and write to the target of the symlink.

This can be exploited to cause corruption to files on the Forgejo server or container that the server process has write access to.

In specific configurations, it is possible to exploit this vulnerability to gain remote shell access to a Forgejo server.  Specifically, remote shell access can be achieved if:
- The paths of sensitive files are known or guessable to an attacker,
- Git access is available through ssh,
- Forgejo provides ssh access through an `authorized_keys` file, which means:
    - The internal ssh server is not used; `[server].START_SSH_SERVER=false`, which is a default.
    - Forgejo manages an `authorized_keys` file; `[server].SSH_CREATE_AUTHORIZED_KEYS_FILE=true`, which is a default.
- And, an attacker can craft the necessary repositories.

The creation of repositories based upon template repositories is being fixed by sandboxing the file access to within the newly created repository, preventing any known symlink or path traversal attack.


### Vulnerability (**Medium**): prevent .forgejo/template from being out-of-repo content
- [CVSS 6.9 Medium](https://www.first.org/cvss/calculator/4-0#CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:N/VI:N/VA:L/SC:N/SI:N/SA:L)
- [Patch](https://codeberg.org/forgejo/forgejo/commit/afbf1efe028a33fb79c1af208d28a993db8ca2af)

When creating a repository based upon a template repository, Forgejo reads the contents of the `.forgejo/template` (or `.gitea/template`) file in the repository in order to identify which files are to be templated content.  In the event that the `.forgejo/template` file was a symlink to a file outside of the repository, Forgejo would follow this symlink to read this file.  This can be exploited to cause resource exhaustion in the Forgejo server, affecting service availability.

This issue is fixed by sandboxing the file access to within the newly created repository, preventing any known symlink or path traversal attack.


### Vulnerability (**Medium**): return on error if an LFS token cannot be parsed
- [CVSS 6.3 Medium](https://www.first.org/cvss/calculator/4-0#CVSS:4.0/AV:N/AC:H/AT:N/PR:N/UI:N/VC:L/VI:N/VA:N/SC:N/SI:N/SA:N)
- [Patch](https://codeberg.org/forgejo/forgejo/commit/fa1a2ba669301238cf3da6a3e746912d76e47f36)

If [LFS](https://git-lfs.com/) is enabled on a Forgejo instance with `[server].LFS_START_SERVER = true` (this is not the default), it was possible for a user to download LFS files for which the OID is known in advance from a private repository to which they did not have read access. This is fixed by returning on error in case the LFS token is invalid instead of returning an inconsistent state.



### Vulnerability (**Low**): prevent commit API from leaking user's hidden email address on valid GPG signed commits
- [CVSS 2.3 Low](https://www.first.org/cvss/calculator/4-0#CVSS:4.0/AV:N/AC:L/AT:P/PR:N/UI:P/VC:L/VI:N/VA:N/SC:N/SI:N/SA:N)
- [Patch](https://codeberg.org/forgejo/forgejo/commit/8885844e72fedc3585a1f42f16c430ab0cbeb90f)

When a signed GPG commit is accessed through the `/repos/{owner}/{repo}/git/commits/{sha}` API, the verified commit's author's primary email address is exposed in the `commit.verification.signer.email` field in violation of the author's desire to keep their email address private.

This has been fixed by returning the signature’s identity, rather than the account's private email address, which is consistent with what is stored in the Git repo and returned by `git log`.


### Go 1.25 Upgrade

To provide both the highest confidence in the security of these fixes, as well as to maintain full backwards compatibility with the current Forgejo behaviour, an upgrade to go 1.25 was required in order to use their [path-traversal resistant `os.Root` APIs](https://go.dev/blog/osroot) which had new APIs added in [go 1.25](https://go.dev/doc/go1.25#ospkgos).

We recognize that this is an unfortunate change to make in security patches for major versions, and could be difficult for Forgejo packagers to adapt to.

If necessary, a go 1.24 implementation of these security fixes is also included in this release, and `go.mod` can be patched to use go 1.24.  The go 1.24 implementation carries the same security guarantees, but introduces two known behaviour changes.  These are edge cases in repository template expansion which are unlikely to affect end-users, but unfortunately could not be addressed safely in go 1.24:

- In the event that a template repo has a file path with a substitution in it, and the file mode is executable, the new file in the target repo will not be executable (and the executable file gets removed).
- In the event that a template repo has a file path with a substitution in it, and the file is a symlink to another in-repo file, the old behaviour would have been to write through the symlink to the file and then rename the symlink; the new behaviour is to write a regular file in-place of the symlink and delete the symlink.




<!--start release-notes-assistant-->

## Release notes
<!--URL:https://codeberg.org/forgejo/forgejo-->
- Security bug fixes
  - [PR](https://codeberg.org/forgejo/forgejo/pulls/9849): <!--number 9849 --><!--line 0 --><!--description W2NvbW1pdF0oaHR0cHM6Ly9jb2RlYmVyZy5vcmcvZm9yZ2Vqby9mb3JnZWpvL2NvbW1pdC84ODg1ODQ0ZTcyZmVkYzM1ODVhMWY0MmYxNmM0MzBhYjBjYmViOTBmKSBwcmV2ZW50IGNvbW1pdCBBUEkgZnJvbSBsZWFraW5nIHVzZXIncyBoaWRkZW4gZW1haWwgYWRkcmVzcyBvbiB2YWxpZCBHUEcgc2lnbmVkIGNvbW1pdHM=-->[commit](https://codeberg.org/forgejo/forgejo/commit/8885844e72fedc3585a1f42f16c430ab0cbeb90f) prevent commit API from leaking user's hidden email address on valid GPG signed commits<!--description-->
  - [PR](https://codeberg.org/forgejo/forgejo/pulls/9849): <!--number 9849 --><!--line 1 --><!--description W2NvbW1pdF0oaHR0cHM6Ly9jb2RlYmVyZy5vcmcvZm9yZ2Vqby9mb3JnZWpvL2NvbW1pdC80NDliNWJmMTBlM2JhNGJhZjJlZjMzYjUzNWI1ZThiYTE1NzExODM4KSBwcmV2ZW50IHdyaXRpbmcgdG8gb3V0LW9mLXJlcG8gc3ltbGluayBkZXN0aW5hdGlvbnMgd2hpbGUgZXZhbHVhdGluZyB0ZW1wbGF0ZSByZXBvcw==-->[commit](https://codeberg.org/forgejo/forgejo/commit/449b5bf10e3ba4baf2ef33b535b5e8ba15711838) prevent writing to out-of-repo symlink destinations while evaluating template repos<!--description-->
  - [PR](https://codeberg.org/forgejo/forgejo/pulls/9849): <!--number 9849 --><!--line 2 --><!--description W2NvbW1pdF0oaHR0cHM6Ly9jb2RlYmVyZy5vcmcvZm9yZ2Vqby9mb3JnZWpvL2NvbW1pdC9hZmJmMWVmZTAyOGEzM2ZiNzljMWFmMjA4ZDI4YTk5M2RiOGNhMmFmKSBwcmV2ZW50IC5mb3JnZWpvL3RlbXBsYXRlIGZyb20gYmVpbmcgb3V0LW9mLXJlcG8gY29udGVudA==-->[commit](https://codeberg.org/forgejo/forgejo/commit/afbf1efe028a33fb79c1af208d28a993db8ca2af) prevent .forgejo/template from being out-of-repo content<!--description-->
  - [PR](https://codeberg.org/forgejo/forgejo/pulls/9849): <!--number 9849 --><!--line 3 --><!--description W2NvbW1pdF0oaHR0cHM6Ly9jb2RlYmVyZy5vcmcvZm9yZ2Vqby9mb3JnZWpvL2NvbW1pdC9mYTFhMmJhNjY5MzAxMjM4Y2YzZGE2YTNlNzQ2OTEyZDc2ZTQ3ZjM2KSByZXR1cm4gb24gZXJyb3IgaWYgYW4gTEZTIHRva2VuIGNhbm5vdCBiZSBwYXJzZWQ=-->[commit](https://codeberg.org/forgejo/forgejo/commit/fa1a2ba669301238cf3da6a3e746912d76e47f36) return on error if an LFS token cannot be parsed<!--description-->
- Localization
  - Updates from Codeberg Translate: [#9825](https://codeberg.org/forgejo/forgejo/pulls/9825) (backport of [#9597](https://codeberg.org/forgejo/forgejo/pulls/9597), [#9696](https://codeberg.org/forgejo/forgejo/pulls/9696))
- Bug fixes
  - [PR](https://codeberg.org/forgejo/forgejo/pulls/9795): <!--number 9795 --><!--line 0 --><!--description Zml4KHBlcmYpOiBhZGQgbWlzc2luZyBpbmRleCBvbiBhY3Rpb25fdGFzayB0YWJsZQ==-->fix(perf): add missing index on action_task table<!--description-->
  - [PR](https://codeberg.org/forgejo/forgejo/pulls/9776): <!--number 9776 --><!--line 0 --><!--description Zml4OiBzdHJpY3QgZXJyb3IgaGFuZGxpbmcgb24gY29ycnVwdGVkIERCIG1pZ3JhdGlvbiB0cmFja2luZyB0YWJsZXM=-->fix: strict error handling on corrupted DB migration tracking tables<!--description-->
  - [PR](https://codeberg.org/forgejo/forgejo/pulls/9764) ([backported](https://codeberg.org/forgejo/forgejo/pulls/9772)): <!--number 9772 --><!--line 0 --><!--description Zml4OiBHTE9CQUxfVFdPX0ZBQ1RPUl9SRVFVSVJFTUVOVCBhbGwgcHJldmVudHMgYWN0aW9ucy9jaGVja291dCBmcm9tIGNsb25pbmcgcmVwb3NpdG9yaWVz-->fix: GLOBAL_TWO_FACTOR_REQUIREMENT all prevents actions/checkout from cloning repositories<!--description-->
- Included for completeness but not user-facing (chores, etc.)
  - [PR](https://codeberg.org/forgejo/forgejo/pulls/9822) ([backported](https://codeberg.org/forgejo/forgejo/pulls/9827)): <!--number 9827 --><!--line 0 --><!--description Y2hvcmU6IHVwZGF0ZSBnbyB0YXJnZXQgbGFuZ3VhZ2UgdmVyc2lvbiB0byB2MS4yNS4w-->chore: update go target language version to v1.25.0<!--description-->
  - [PR](https://codeberg.org/forgejo/forgejo/pulls/9824): <!--number 9824 --><!--line 0 --><!--description VXBkYXRlIGRhdGEuZm9yZ2Vqby5vcmcvb2NpL2dvbGFuZyBEb2NrZXIgdGFnIHRvIHYxLjI1ICh2MTMuMC9mb3JnZWpvKQ==-->Update data.forgejo.org/oci/golang Docker tag to v1.25 (v13.0/forgejo)<!--description-->
  - [PR](https://codeberg.org/forgejo/forgejo/pulls/9816): <!--number 9816 --><!--line 0 --><!--description VXBkYXRlIGRlcGVuZGVuY3kgZ28gdG8gdjEuMjUgKHYxMy4wL2Zvcmdlam8p-->Update dependency go to v1.25 (v13.0/forgejo)<!--description-->
<!--end release-notes-assistant-->
