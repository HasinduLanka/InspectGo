<script lang="ts">
  let txtURL = "https://go.dev/";
  let report: any;

  let Inspect = async (url: string) => {
    let reportResp = await fetch("/api/inspect", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url: url }),
    });

    report = await reportResp.text();
  };
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
