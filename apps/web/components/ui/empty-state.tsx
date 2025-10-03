import React from 'react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { LucideIcon } from 'lucide-react';

interface EmptyStateProps {
  icon: LucideIcon;
  title: string;
  description: string;
  action?: {
    label: string;
    onClick: () => void;
    variant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link";
  };
  variant?: "default" | "minimal";
  className?: string;
}

export function EmptyState({ 
  icon: Icon, 
  title, 
  description, 
  action, 
  variant = "default",
  className 
}: EmptyStateProps) {
  return (
    <div className={cn(
      "flex flex-col items-center justify-center text-center py-12 px-6",
      className
    )}>
      <div className="relative mb-6">
        <div className="w-14 h-14 rounded-full bg-muted/50 flex items-center justify-center">
          <Icon className="w-7 h-7 text-muted-foreground/60" />
        </div>
        {variant === "default" && (
          <div className="absolute inset-0 w-14 h-14 rounded-full bg-gradient-to-br from-primary/5 to-primary/10 animate-pulse" />
        )}
      </div>
      
      <h3 className="font-semibold text-base text-foreground mb-2">
        {title}
      </h3>
      
      <p className="text-muted-foreground text-sm max-w-sm leading-relaxed mb-6">
        {description}
      </p>
      
      {action && (
        <Button 
          onClick={action.onClick}
          variant={action.variant || "default"}
          size="sm"
          className="min-w-[120px]"
        >
          {action.label}
        </Button>
      )}
    </div>
  );
}
