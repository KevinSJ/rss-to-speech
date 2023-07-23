# RSS to Speech

---

### Why?

- I want to try out a personal project written completely in Go language.
- I like to use RSS, and content are lengthy at times.
- Google has a free quota of 3 **million** characters for standard voice and 1 **million** characters for wavenet voice **_per month_**, why not make use of it.

### Technology

- Go
- Google Text-to-speech API

### Usage
- Rename config.yaml.example to config.yaml, fill in the relevant configurations.
- You will need to download the Google credential and put it at the current directory (or somewhere else if you change the credential path)
- Add RSS feeds (has to emit full text, for now)
- Build the program from source `make`
- Run the program, `./rss-to-speech`
- Once finished, you will see different folders with title of the RSS feed as name.
