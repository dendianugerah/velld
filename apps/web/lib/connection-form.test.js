import { describe, expect, test } from "bun:test";
import { connectionFormValues, d1ConnectionPayload, parseConnectionUrl } from "./connection-form";

describe("connection URL submission", () => {
  test("returns decoded PostgreSQL credentials and TLS preference for the current submission", () => {
    expect(parseConnectionUrl("postgres://backup%40team:p%3A%40ss@db.internal:5433/app%20data?sslmode=disable")).toEqual({
      type: "postgresql", host: "db.internal", port: 5433, username: "backup@team", password: "p:@ss", database: "app data", mongo_auth_source: "", ssl: false,
    });
  });
  test("preserves MongoDB authentication database and tls=false", () => {
    expect(parseConnectionUrl("mongodb://user:pass@localhost/app?authSource=admin&tls=false")).toMatchObject({ type: "mongodb", port: 27017, mongo_auth_source: "admin", database: "app", ssl: false });
  });
  test("supports IPv6 without URL brackets in the connection host", () => {
    expect(parseConnectionUrl("mysql://user:pass@[::1]:3307/app")).toMatchObject({ host: "::1", port: 3307, ssl: true });
  });
  test("rejects unsupported and malformed URLs without repeating their credentials", () => {
    for (const value of ["https://user:private@host/db", "postgres://", "postgres://user:%ZZ@host/db", "mysql://user:private@host:99999/db"]) {
      expect(() => parseConnectionUrl(value)).toThrow();
      try { parseConnectionUrl(value); } catch (error) { expect(error.message).not.toContain("private"); }
    }
  });
});

test("D1 submissions exclude SQL credentials, and editing does not require the stored token", () => {
  const values = connectionFormValues({ type: "d1", name: " D1 backup ", host: "api.cloudflare.com", d1_account_id: " account ", d1_database_id: " database " });
  const payload = d1ConnectionPayload({ ...values, password: "sql-only", ssh_enabled: true, ssh_private_key: "ssh-only" });
  expect(payload).toMatchObject({ name: "D1 backup", type: "d1", d1_account_id: "account", d1_database_id: "database", password: "" });
  expect(payload.ssh_private_key).toBeUndefined();
  expect(payload.d1_api_token).toBeUndefined();
  expect(d1ConnectionPayload({ ...values, d1_api_token: " new-token " }).d1_api_token).toBe("new-token");
});
