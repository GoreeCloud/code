import assert from "node:assert/strict";
import { createServer } from "node:http";
import { after, before, test } from "node:test";
import { ForgejoProvider } from "./index.ts";

let baseUrl = "";
let server: ReturnType<typeof createServer>;

before(async () => {
  server = createServer((request, response) => {
    response.setHeader("content-type", "application/json");
    const path = request.url ?? "";
    if (path === "/api/v1/version") return response.end(JSON.stringify({ version: "11.0.0-test" }));
    if (path.startsWith("/api/v1/user/repos")) return response.end(JSON.stringify([{ id: 7, name: "code", owner: { login: "goreecloud" }, default_branch: "main", private: true, html_url: "https://forge.test/goreecloud/code" }]));
    if (path === "/api/v1/repos/goreecloud/code/branches?limit=100") return response.end(JSON.stringify([{ name: "main", protected: true, commit: { id: "abc123" } }]));
    if (path === "/api/v1/repos/goreecloud/code/issues?state=all&limit=50") return response.end(JSON.stringify([{ number: 1, title: "Issue", state: "open", html_url: "https://forge.test/i/1" }, { number: 2, title: "PR", state: "open", pull_request: {}, html_url: "https://forge.test/p/2" }]));
    response.statusCode = 404;
    response.end(JSON.stringify({ message: "not found" }));
  });
  await new Promise<void>((resolve) => server.listen(0, "127.0.0.1", resolve));
  const address = server.address();
  if (!address || typeof address === "string") throw new Error("Mock server failed to bind");
  baseUrl = `http://127.0.0.1:${address.port}`;
});

after(async () => new Promise<void>((resolve, reject) => server.close((error) => error ? reject(error) : resolve())));

test("reports provider health and version", async () => {
  const provider = new ForgejoProvider({ baseUrl, token: "test-token" });
  const health = await provider.health();
  assert.equal(health.ok, true);
  assert.equal(health.provider, "forgejo");
  assert.equal(health.version, "11.0.0-test");
});

test("maps repositories and branches into provider-neutral contracts", async () => {
  const provider = new ForgejoProvider({ baseUrl, token: "test-token" });
  const [repository] = await provider.repositories();
  assert.deepEqual(repository, {
    id: "7", owner: "goreecloud", name: "code",
    defaultBranch: "main", private: true, webUrl: "https://forge.test/goreecloud/code",
  });
  assert.deepEqual(await provider.branches({ owner: "goreecloud", name: "code" }), [{ name: "main", sha: "abc123", protected: true }]);
});

test("filters pull requests from issue results", async () => {
  const provider = new ForgejoProvider({ baseUrl, token: "test-token" });
  const issues = await provider.issues({ owner: "goreecloud", name: "code" });
  assert.equal(issues.length, 1);
  assert.equal(issues[0].number, 1);
});

test("requires an explicit discovery identity", async () => {
  const provider = new ForgejoProvider({ baseUrl });
  await assert.rejects(() => provider.repositories(), /FORGEJO_TOKEN or FORGEJO_USERNAME/);
});
