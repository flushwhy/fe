# Chocolatey packaging

The release workflow builds the Windows binary and packages it as a Chocolatey package.
On each `vX.Y.Z` tag, it uploads `fe.X.Y.Z.nupkg` to the GitHub Release. If the
repository secret `CHOCOLATEY_API_KEY` is configured, the workflow also pushes the
package to Chocolatey Community Repository.

To enable publishing:
1. Create an account at https://community.chocolatey.org/.
2. Obtain your API key from your Chocolatey account.
3. Add it to the GitHub repository as Actions secret `CHOCOLATEY_API_KEY`.

The package downloads the matching Windows executable from the GitHub Release and
verifies its SHA-256 checksum before adding the `fe` command to PATH.

For a manual package build, substitute `__VERSION__` and `__SHA256__` in these
files, then run `choco pack fe.nuspec` from the packaging directory.
