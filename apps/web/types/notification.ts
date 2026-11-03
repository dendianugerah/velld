import { Base } from "./base";

export type NotificationType =
    | 'backup_failed'
    | 'backup_completed'
    | 'verification_failed'
    | 'wal_gap_detected';
export type NotificationStatus = 'read' | 'unread';

export interface Notification {
    id: string;
    title: string;
    message: string;
    type: NotificationType;
    status: NotificationStatus;
    created_at: string;
    metadata?: {
        connection_id?: string;
        error?: string;
        [key: string]: any;
    };
}

export type NotificationResponse = Base<Notification[]>;