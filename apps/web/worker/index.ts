// Serves the static marketing site; canonicalizes www -> apex (D7).
interface Env {
  ASSETS: Fetcher;
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    if (url.hostname === "www.servienta.com") {
      url.hostname = "servienta.com";
      return Response.redirect(url.toString(), 301);
    }
    // Clean /docs URL -> the docs page asset.
    if (url.pathname === "/docs" || url.pathname === "/docs/") {
      url.pathname = "/docs.html";
      return env.ASSETS.fetch(new Request(url.toString(), request));
    }
    if (url.pathname === "/demo" || url.pathname === "/demo/") {
      url.pathname = "/demo.html";
      return env.ASSETS.fetch(new Request(url.toString(), request));
    }
    return env.ASSETS.fetch(request);
  },
} satisfies ExportedHandler<Env>;
