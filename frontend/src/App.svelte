<script lang="ts">
  import type { InspectResponse } from "./Types";
  import { tweened } from "svelte/motion";
  import { cubicOut } from "svelte/easing";

  let txtURL: string = "https://go.dev/";
  let report: InspectResponse;
  let isLoading: boolean = false;

  async function streamResponse(response: Response) {
    const reader = response.body
      .pipeThrough(new TextDecoderStream())
      .getReader();

    let received = "";
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      received += value;

      if (received) {
        try {
          report = JSON.parse(received);
          received = "";
          console.log("Report stream received");
        } catch (error) {
          console.log("Partially received report stream");
        }
      }
    }
    console.log("Report fully received");
  }

  let inspectLock: boolean = false;

  async function Inspect(url: string) {
    if (inspectLock) return;

    report = null;
    inspectLock = true;

    setTimeout(() => {
      inspectLock = false;
    }, 2000);

    isLoading = true;

    let resp = await fetch("/api/inspect", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "inspecter-response-streamable": "true",
      },
      body: JSON.stringify({ url: url }),
    });
    streamResponse(resp);
  }

  function getColor(code: number) {
    let color: string;

    if (code == 0) {
      color = "#4b5563";
    } else if (code < 300) {
      color = "#15803d";
    } else if (code < 400) {
      color = "#a16207";
    } else {
      color = "#881337";
    }

    return color;
  }

  const progress = tweened(0, {
    duration: 3000,
    easing: cubicOut,
  });

  $: if (report && report.total_link_count) {
    let totalAnalysed =
      report.total_link_count - report.not_analysed_link_count;

    let fraction = (totalAnalysed / report.total_link_count) * 0.8 + 0.2;

    progress.set(fraction);

    if (report.not_analysed_link_count == 0) {
      isLoading = false;
    }
  }
</script>

<div class="container">
  <center style="margin-top:25px;">
    <h1 style="margin: 10px">Inspect Go Dev</h1>
    <p style="margin: 10px">Inspect webpages and see how they are built.</p>
    <div
      style="height: 100px; border: 3px solid #111827;background-color:#164e63; margin-top:25px;border-radius:10px; padding: 20px 10px"
    >
      <p>Enter URL to inspect:</p>
      <br />
      <input
        type="text"
        style="width: 30rem; text-align: center; "
        bind:value={txtURL}
      />
      {#if txtURL}
        <button on:click={() => Inspect(txtURL)}>Inspect</button>
      {/if}
    </div>
  </center>

  {#if report}
    {#if report.status_code < 300}
      <center style="margin-top:25px;">
        <div
          style="border: 3px solid #111827;background-color:#164e63; margin-top:25px;border-radius:10px; padding: 20px 10px"
        >
          {#if isLoading}
            <p>Links are still being analyzed</p>
          {:else}
            <p>Analysis complete</p>
          {/if}

          <progress value={$progress} />
        </div>
      </center>
      <div
        style="border: 3px solid #111827; background-color:#164e63; margin-top:25px; border-radius:10px; padding: 20px 2rem;"
      >
        <h1 style="text-align:center">Inspect Report</h1>
        <br />
        <table style="font-size:1.2rem; text-align-left; width:100%">
          <colgroup>
            <col style="width: 15rem" />
            <col style="width: 1rem" />
            <col style="width: auto" />
          </colgroup>
          <tr>
            <td>URL</td>
            <td>:</td>
            <td>{report.url}</td>
          </tr>
          <tr>
            <td>Page Title</td>
            <td>:</td>
            <td>{report.page_title}</td>
          </tr>
          <tr>
            <td>HTML Version</td>
            <td>:</td>
            <td>{report.html_version}</td>
          </tr>
          <tr>
            <td>Status Code </td>
            <td>:</td>
            <td>{report.status_code}</td>
          </tr>
          <tr>
            <td>Status Message</td>
            <td>:</td>
            <td>{report.status_msg}</td>
          </tr>
          <tr>
            <td colspan="3"><hr /></td>
          </tr>
          <tr>
            <td>Total Links</td>
            <td>:</td>
            <td>{report.total_link_count}</td>
          </tr>
          <tr>
            <td>Accessible Links</td>
            <td>:</td>
            <td>{report.accessible_link_count}</td>
          </tr>
          <tr>
            <td>Inaccessible Links</td>
            <td>:</td>
            <td>{report.inaccessible_link_count}</td>
          </tr>
          <tr>
            <td>Non web Links</td>
            <td>:</td>
            <td>
              {report.total_link_count -
                report.accessible_link_count -
                report.inaccessible_link_count -
                report.not_analysed_link_count}
            </td>
          </tr>
          <tr>
            <td>Links not yet analysed</td>
            <td>:</td>
            <td>{report.not_analysed_link_count}</td>
          </tr>
          <tr>
            <td>External Links</td>
            <td>:</td>
            <td>{report.external_link_count}</td>
          </tr>
          <tr>
            <td>Internal Links</td>
            <td>:</td>
            <td>{report.internal_link_count}</td>
          </tr>
          <tr>
            <td colspan="3"><hr /></td>
          </tr>
          <tr>
            <td>Heading Types</td>
            <td>:</td>
            <td>{Object.keys(report.headings).length}</td>
          </tr>
          <tr>
            <td colspan="3"><hr /></td>
          </tr>
        </table>
        <br />
        <h2 style="text-align:center;">Headings</h2>

        {#each Object.keys(report.headings) as head, headindex}
          <div
            style="border: 3px solid #111827; margin-top:25px; border-radius:10px; padding: 20px 10px; margin-bottom:20px; background-color:#1e3a8a"
          >
            <table style="font-size:1.2rem; text-align-left; width:100%;">
              <colgroup>
                <col style="width: 150px" />
                <col style="width: auto" />
              </colgroup>
              <tr>
                <td
                  style="border-right: 1px solid gray; text-align: center; padding: 20px; "
                >
                  <h1>{head}</h1>
                  <p>
                    {report.headings[head].length}
                    {report.headings[head].length == 1 ? "heading" : "headings"}
                  </p>
                </td>
                <td>
                  <div>
                    <table
                      style="font-size:1.2rem; text-align-left; width:100%;"
                    >
                      {#each report.headings[head] as h, hindex}
                        <tr>
                          <td style="border-bottom: 1px solid gray;">
                            <p>{h}</p>
                          </td>
                        </tr>
                      {/each}
                    </table>
                  </div>
                </td>
              </tr>
            </table>
          </div>
        {/each}
        <br />
        <hr />
        <br />
        <h2
          style="font-weight: bold ; width:100%; text-align:center; margin-top:20px"
        >
          Links
        </h2>

        {#each report.links as link, linkindex}
          <div
            style={`border: 3px solid #111827; margin-top:25px;border-radius:10px; padding: 20px 10px; margin-bottom:20px; 
          background-color:${getColor(link.status_code)}`}
          >
            <table style="font-size:1.2rem; text-align-left; width:100%;">
              <colgroup>
                <col style="width: 150px" />
                <col style="width: 20px" />
                <col style="width: auto" />
              </colgroup>
              <tr>
                <td>Link</td>
                <td>:</td>
                <td>
                  <a href={link.url} style="color: white;">{link.url}</a>
                </td>
              </tr>
              <tr>
                <td>Type </td>
                <td>:</td>
                <td>{link.type}</td>
              </tr>
              {#if link.text}
                <tr>
                  <td>Text</td>
                  <td>:</td>
                  <td>{link.text}</td>
                </tr>
              {/if}
              {#if link.status_code}
                <tr>
                  <td>Status Code </td>
                  <td>:</td>
                  <td>{link.status_code}</td>
                </tr>
              {/if}
            </table>
          </div>
        {/each}
      </div>
    {:else}
      <div
        style="border: 3px solid #7f1d1d; background-color:#164e63; margin-top:25px; border-radius:10px; padding: 20px 2rem;"
      >
        <h1 style="text-align:center">This webpage cannot be reached</h1>
        <br />
        <table style="font-size:1.2rem; text-align-left; width:100%">
          <colgroup>
            <col style="width: 15rem" />
            <col style="width: 1rem" />
            <col style="width: auto" />
          </colgroup>
          <tr>
            <td>URL</td>
            <td>:</td>
            <td>{report.url}</td>
          </tr>
          <tr>
            <td>Status Code </td>
            <td>:</td>
            <td>{report.status_code}</td>
          </tr>
          <tr>
            <td>Status Message</td>
            <td>:</td>
            <td>{report.status_msg}</td>
          </tr>
        </table>
      </div>
    {/if}
  {:else if isLoading}
    <h1 style="text-align: center;">Requesting...</h1>
  {/if}
</div>

<style>
  .container {
    width: 100%;
    max-width: 1200px;
    margin: 0 auto;
  }
  input {
    border: 3px solid #111827;
    border-radius: 5px;
    background-color: #334155;
    color: #f8fafc;
    font-size: 1.2em;
    padding: 5px 15px;
    margin: 5px;
  }
  input:focus {
    background-color: #64748b;
  }
  button {
    border: 3px solid #111827;
    border-radius: 5px;
    background-color: #64748b;
    color: #f8fafc;
    font-size: 1.2em;
    padding: 5px 15px;
    margin: 5px;
    cursor: pointer;
  }
  button:hover {
    background-color: #334155;
  }
  button:active {
    background-color: #111827;
  }
  td {
    padding: 5px 10px;
    text-align: left;
  }
</style>
