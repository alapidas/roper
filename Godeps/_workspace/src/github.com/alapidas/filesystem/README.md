# filesystem
[![Circle CI](https://circleci.com/gh/alapidas/filesystem.svg?style=svg)](https://circleci.com/gh/alapidas/filesystem)

A simple filesystem interface in Go

The filesystem package provides a unified interface to a real backed filesystem.  An example use case for this might be a server that wants to provide CRUD access to a filesystem.

Here's an example of how you might use this package.

The first step is to create a Filesystemer object based on the absolute filesystem path you want to represent (using a provided constructor):
```go
fs := NewPassThroughFilesystem('/my/abs/path', false)
```
At this point, all interaction with the Filesystemer should be relative to the root path.  For example, to check if /my/abs/path/lower/dir exists, you'd do something like:
```go
fs.PathExists('lower/dir') // false
```
If it did not exist, and you wanted to make a director there, you could do:
```go
fs.MkPath('lower/dir')
```
If you wanted to then make a file in this path, you might do:
```go
fileBytes, err := ioutil.ReadFile('/tmp/roflcopter.jpeg')
fs.MkFile('lower/dir/roflcopter.jpeg', fileBytes)
```
Then to retrieve the file, you'd do:
```go
myFiler, err := fs.GetFile('lower/dir/roflcopter.jpeg')
```
To remove the newly created file, this would work:
```go
fs.RmPath('lower/dir/roflcopter.jpeg')
// Alternatively, recursively delete the whole directory
// fs.RmPath('lower')
```
