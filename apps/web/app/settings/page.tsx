import { Workspace } from "@/components/layout/workspace";
import { ProfileSettings } from "@/components/views/settings/profile-settings";

export default function SettingsPage() {
  return (
    <Workspace title="Settings" description="Manage your account, alerts, and backup storage." section="Account">
      <ProfileSettings />
    </Workspace>
  );
}