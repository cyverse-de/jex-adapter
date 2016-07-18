FROM jeanblanchard/alpine-glibc
ARG git_commit=unknown
ARG buildenv_git_commit=unknown
ARG version=unknown
COPY jex-adapter /bin/jex-adapter
CMD ["jex-adapter" "--help"]
