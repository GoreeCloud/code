import type { ForgeProvider } from "@goreecloud/code-contracts";
import { ForgejoProvider } from "@goreecloud/code-forgejo";
import { createCodeServer } from "./server.js";

const port = numberEnv("PORT", 8787);
const host = process.env.HOST ?? "0.0.0.0";
const provider = createProvider();
const server = createCodeServer(provider, {
  corsOrigin: process.env.CORS_ORIGIN ?? "http://localhost:5173",
});

server.listen(port, host, () => {
  console.log(`GoreeCloud Code API listening on http://${host}:${port}`);
});

function createProvider(): ForgeProvider {
  const baseUrl = requiredEnv("FORGEJO_BASE_URL");
  const token = process.env.FORGEJO_TOKEN;
  const username = process.env.FORGEJO_USERNAME;
  return new ForgejoProvider({
    baseUrl,
    ...(token ? { token } : {}),
    ...(username ? { username } : {}),
    timeoutMs: numberEnv("FORGEJO_TIMEOUT_MS", 10_000),
  });
}

function requiredEnv(name: string): string {
  const value = process.env[name]?.trim();
  if (!value) throw new Error(`${name} is required`);
  return value;
}

function numberEnv(name: string, fallback: number): number {
  const raw = process.env[name];
  if (!raw) return fallback;
  const value = Number(raw);
  if (!Number.isFinite(value) || value <= 0) throw new Error(`${name} must be a positive number`);
  return value;
}
