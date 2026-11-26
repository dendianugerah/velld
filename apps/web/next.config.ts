import type { NextConfig } from "next";
import { dirname, resolve } from "node:path";
import dotenv from "dotenv";

dotenv.config({ path: "../../.env" });

const nextConfig: NextConfig = {
  /* config options here */
  output: "standalone",
  outputFileTracingRoot: resolve(dirname(require.resolve("next/package.json")), "../.."),
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
  },
};

export default nextConfig;
