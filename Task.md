# Objective

The objective is to build a web application that does an analysis of a web-page/URL.

The application should show a form with a text field in which users can type in the URL of the webpage to be analyzed. Additionally to the form, it should contain a button to send a request to the server.

After processing the results should be shown to the user.

Results should contain next information:

- [x] What HTML version has the document?
- [x] What is the page title?
- [x] How many headings of what level are in the document?
- [x] How many internal and external links are in the document? Are there any inaccessible links and how many?
- [x] Does the page contain a login form?

In case the URL given by the user is not reachable an error message should be presented to a user. The message should contain the HTTP status code and a useful error description.

# Challenges

I faced the following challenges while working on this project.

## Many websites do not allow scanning with tools like this

Many websites block HTTP requests from non-browser clients to prevent scams, making many website links appear broken. I solved this issue by setting custom HTTP headers on requests to disguise as a web browser.

However, this is only used to analyze external links on a web page. Directly scanning a protected website is disabled due to possible legal reasons.

Some websites like LinkedIn, still do not respond well to inspection requests.

## Link analysis taking too long

Even with a 2MB/s connection, analyzing links of an average Wikipedia page takes around 2 minutes because they contain thousands of links. If the user has to wait this amount of time before seeing any result, it will be a bad user experience.

Possible solutions:

- Present report in pieces with HTTP long polling
  - This is the simplest and most common approach.
  - However, long polling sends multiple requests repeatedly. As this project is targeted to run on a serverless platform, these requests can not be guaranteed to hit the same serverless instance. Solving that problem using external caching will make the project too complex.
- Use web sockets to stream the state
  - This is the best option for an industry-standard project. But I didn't select it for its client-side complexity.
- **Stream server state using HTTP response streaming**
  - This solution is lightweight and simple.
  - This could easily go wrong without proper race-condition and memory management. But, with Golang's concurrency model, this is a piece of cake. **Therefore I implemented it.**
  - Server accepts only one HTTP request and keeps it alive.
  - Client makes use of the response streaming features of browser fetch API, updating the user interface every time a response is received.
  - The first report containing the basic information about the webpage is shown almost immediately, making a swift user experience.
  - Sequel responses will be received every 20 seconds, containing the updated state of the report.

# Testing against real websites that changes over time

Writing test cases to test application functionality against real websites is a challenge. Their titles, headings, and links change over time. I solved it using web pages from archive.org instead of live websites for several tests.

## Client-side rendered content

Websites built with frameworks like React or Svelte use client-side rendering where the actual content of a web page is rendered inside the browser. Inspecting these kinds of web pages is not possible unless they are hydrated or pre-rendered on the origin server.

A possible solution is to use a headless browser environment such as puppeteer or selenium to render the web page and inspect it. But this will bring the project complexity to another level. Therefore, this isn't yet solved.

I see this as a great future improvement.
