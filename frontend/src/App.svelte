<script lang="ts">
  let txtURL: string = "https://go.dev/";
  let report: string;
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
          report = JSON.stringify(JSON.parse(received), null, "  ");
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
</script>

<main style="margin: 4rem;">
  <div style=" text-align: center;">
    <h1>Inspect Go</h1>
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

  {#if report}
    <div>
      <h2>Report</h2>
      <pre style="white-space: pre-wrap;"> {report} </pre>
    </div>
  {/if}
</main>
