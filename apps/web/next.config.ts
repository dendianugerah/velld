import type { NextConfig } from "next";
import dotenv from "dotenv";

dotenv.config({ path: "../../.env" });

const nextConfig: NextConfig = {
  /* config options here */
  output: "standalone",
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
    ALLOW_REGISTER: process.env.ALLOW_REGISTER,
  },
};

export default nextConfig;
