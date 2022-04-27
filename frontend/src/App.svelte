<script lang="ts">
  let msg;

  let init = async () => {
    let helloResp = await fetch("/api/hello");
    msg = await helloResp.text();
  };

  init();
</script>

<main style="margin: 4rem; text-align: center;">
  <h1>
    Project template for developing web applications with Svelte front-end and
    Golang server-less back-end, hosted on Vercel.
  </h1>
  <pre> {msg} </pre>
</main>
