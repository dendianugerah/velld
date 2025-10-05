import { Bell, Check } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { formatDistanceToNow, parseISO } from "date-fns";
import { Notification } from "@/types/notification";
import { cn } from "@/lib/utils";

interface NotificationSidebarProps {
  notifications: Notification[];
  isLoading: boolean;
  onMarkAsRead: (ids: string) => void;
}

export function NotificationSidebar({ notifications, isLoading, onMarkAsRead }: NotificationSidebarProps) {
  const unreadNotifications = notifications.filter(n => n.status === 'unread');
  const recentNotifications = notifications.slice(0, 5);

  return (
    <div className="w-full lg:w-80 lg:border-l lg:pl-4">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <h4 className="text-sm font-semibold">Recent Notifications</h4>
          {unreadNotifications.length > 0 && (
            <Badge variant="secondary" className="bg-primary/10 text-primary">
              {unreadNotifications.length}
            </Badge>
          )}
        </div>
      </div>
      <ScrollArea className="h-[300px] lg:h-[400px]">
        <div className="space-y-3 pr-4">
          {isLoading ? (
            Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="p-2 rounded-lg border animate-pulse">
                <div className="h-4 bg-muted rounded w-3/4 mb-2"></div>
                <div className="h-3 bg-muted rounded w-1/2"></div>
              </div>
            ))
          ) : recentNotifications.length === 0 ? (
            <div className="text-center text-sm text-muted-foreground py-4">
              No notifications
            </div>
          ) : (
            recentNotifications.map((notification) => (
              <div
                key={notification.id}
                onClick={() => {
                  if (notification.status === 'unread') {
                    onMarkAsRead(notification.id);
                  }
                }}
                className={cn(
                  "p-3 rounded-lg border transition-all",
                  "hover:shadow-sm",
                  notification.status === 'unread' 
                    ? "bg-background hover:bg-accent/50 cursor-pointer border-primary/20" 
                    : "bg-muted/30",
                )}
              >
                <div className="flex gap-4">
                  <div className={cn(
                    "p-2 rounded-md self-start",
                    notification.status === 'unread' 
                      ? "bg-primary/10" 
                      : "bg-muted"
                  )}>
                    {notification.status === 'unread' 
                      ? <Bell className="h-4 w-4 text-primary" />
                      : <Check className="h-4 w-4 text-muted-foreground" />
                    }
                  </div>
                  <div className="flex-1 min-w-0">
                    <h5 className={cn(
                      "text-sm truncate",
                      notification.status === 'unread' ? "font-medium" : "text-muted-foreground"
                    )}>
                      {notification.title}
                    </h5>
                    <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                      {notification.message}
                    </p>
                    <div className="flex items-center justify-between mt-2">
                      <p className="text-xs text-muted-foreground">
                        {formatDistanceToNow(parseISO(notification.created_at), { addSuffix: true })}
                      </p>
                      {notification.status === 'unread' && (
                        <span className="h-2 w-2 rounded-full bg-primary"/>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </ScrollArea>
    </div>
  );
}
