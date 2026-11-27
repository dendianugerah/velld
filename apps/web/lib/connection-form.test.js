import { describe, expect, test } from "bun:test";
import { parseConnectionUrl } from "./connection-form";

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
