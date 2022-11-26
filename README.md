winrmcp
=======

Copy files to a remote host using WinRM

_This project is not being actively worked on._ See issues:

- [Seeking maintainers](https://github.com/packer-community/winrmcp/issues/38)
- [Sunset winrmcp](https://github.com/packer-community/winrmcp/issues/39)

---

Example:

    make
    bin/winrmcp -help
    bin/winrmcp -user=vagrant -pass=vagrant ~/Downloads/fortune.jpg C:/Cookies/fortune.jpg
