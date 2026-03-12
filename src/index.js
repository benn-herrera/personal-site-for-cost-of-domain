export default {
  async fetch(request, env) {
    // each entry in this list needs a directory in public/ of the same name to be the content root for that site.
    const sites = ["my-personal-site.me", "my-other-personal-site.me"];
    const url = new URL(request.url);
    var domain = url.hostname;
    let setCookie = null;

    // we just want the bare domain name
    if (domain.startsWith("www.")) {
      domain = domain.substring(4);
    }

    // when request is direct to worker allow ?d=<entry-from-sites>
    // for checking multiplexing and seeing desired content root.
    // ?d param sets a session cookie so internal links keep working.
    if (
      url.hostname == "localhost" ||
      url.hostname.startsWith("192.") ||
      url.hostname.startsWith("127.") ||
      url.hostname.endsWith("workers.dev")
    ) {
      const paramD = url.searchParams.get("d");
      if (paramD) {
        domain = paramD;
        // cookie value must be plain alphanum — read side parses via URLSearchParams
        setCookie = `d=${paramD}; Path=/; SameSite=Lax`;
      } else {
        const cookies = new URLSearchParams(
          (request.headers.get("Cookie") || "").replaceAll("; ", "&"),
        );
        domain = cookies.get("d") || sites[0];
      }
    }

    // site not handled
    if (!sites.includes(domain)) {
      return new Response("Service Unavailable", { status: 503 });
    }

    var pathname = url.pathname;

    // clean RSS alias
    if (["/feed/", "/feed"].includes(pathname)) {
      pathname = "/feed.xml";
    }

    // Map the internal asset URL to the domain-specific folder
    const assetUrl = new URL(`/${domain}${pathname}`, url.origin);

    // Passing the original 'request' ensures the Method and Headers
    // are preserved for Cloudflare's internal processing.
    const response = await env.ASSETS.fetch(new Request(assetUrl, request));

    if (setCookie) {
      const r = new Response(response.body, response);
      r.headers.append("Set-Cookie", setCookie);
      return r;
    }
    return response;
  },
};
