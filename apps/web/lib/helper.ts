export const isAuthenticated = () => {
  if (typeof window === "undefined") return false;
  return !!localStorage.getItem("token");
};

export const getToken = () => {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("token");
};

export const logout = () => {
  if (typeof window === "undefined") return;
  localStorage.removeItem("token");
  window.location.href = "/login";
};

export function formatSize(bytes: number): string {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

export function calculateDuration(start: string, end: string) {
  const startTime = new Date(start);
  const endTime = new Date(end);
  const durationMs = endTime.getTime() - startTime.getTime();
  const minutes = Math.floor(durationMs / 60000);
  const seconds = Math.floor((durationMs % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
}

export const getScheduleFrequency = (cronSchedule: string | undefined) => {
  if (!cronSchedule) return "";
  switch (cronSchedule) {
    case "0 */1 * * * *":
      return "test"; // Updated to match 1-minute schedule
    case "0 0 * * * *":
      return "hourly";
    case "0 0 0 * * *":
      return "daily";
    case "0 0 0 * * 0":
      return "weekly";
    case "0 0 0 1 * *":
      return "monthly";
    default:
      return "custom";
  }
};

export const checkIsAllowRegister = (): boolean => {
  const env = process.env.ALLOW_REGISTER;
  if (typeof env === "undefined") return true;

  return env === "true";
};
