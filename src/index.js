function isUndeployed(hostname) {
  return hostname == "localhost" ||
    hostname.startsWith("127.") ||
    hostname.startsWith("192.") ||
    hostname.endsWith("workers.dev");
}

function escape(str) {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
}

function withSetCookieSafe(response, setCookie, responseIsImmutable) {
  if (!setCookie) {
    return response;
  }
  const r = responseIsImmutable ? new Response(response.body, response) : response;
  r.headers.append("Set-Cookie", setCookie);
  return r;
}

export default {
  async fetch(request, env) {
    // `make` from any site content root directory will keep sites list in sync with public/
    // DO NOT touch this line vvv
    const sites = ["bennherrera.dev", "bennherrera.me"];
    // DO NOT touch this line ^^^
    const url = new URL(request.url);
    var domain = url.hostname;
    let setCookie = null;

    // we just want the bare domain name.
    if (domain.startsWith("www.")) {
      domain = domain.substring(4);
    }

    // when request is direct to worker allow ?d=[site-domain.tld]
    // for checking multiplexing and seeing desired content root.
    // ?d param sets a session cookie so internal links keep working.
    if (isUndeployed(url.hostname)) {
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

    if (!sites.includes(domain)) {
      const unavailableUrl = new URL("/503.html", url.origin);
      const unavailable = await env.ASSETS.fetch(
        new Request(unavailableUrl.toString()),
      );
      const html = (await unavailable.text()).replace(
        "<!--DOMAIN-->",
        escape(domain),
      );
      const headers = new Headers(unavailable.headers);
      headers.set("Content-Type", "text/html; charset=utf-8");
      return new Response(html, { status: 503, headers });
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

    // ASSETS redirects (e.g. to add a trailing slash) put the internal
    // /${domain}/ prefix into Location; strip it so the browser is sent
    // to a worker-visible URL rather than re-entering with a doubled prefix.
    if (response.status >= 300 && response.status < 400) {
      const location = response.headers.get("Location");
      const prefix = `/${domain}`;
      if (location && location.startsWith(`${prefix}/`)) {
        const headers = new Headers(response.headers);
        headers.set("Location", location.slice(prefix.length));
        return withSetCookieSafe(new Response(null, {
          status: response.status,
          statusText: response.statusText,
          headers,
        }), setCookie, /*responseIsImmutable=*/ false);
      }
    }

    if (response.status === 404) {
      const notFoundUrl = new URL(`/${domain}/404.html`, url.origin);
      const notFound = await env.ASSETS.fetch(
        new Request(notFoundUrl.toString()),
      );
      const html = (await notFound.text()).replace("<!--PATH-->", escape(url.pathname));
      const headers = new Headers(notFound.headers);
      headers.set("Content-Type", "text/html; charset=utf-8");
      return withSetCookieSafe(new Response(html, { status: 404, headers }), setCookie, /*responseIsImmutable=*/ false);
    }

    return withSetCookieSafe(response, setCookie, /*responseIsImmutable=*/ true);
  },
};
