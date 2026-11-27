interface PageHeaderProps {
  title: string;
  description?: string;
  action?: React.ReactNode;
}

export function PageHeader({ title, description, action }: PageHeaderProps) {
  return <div className="mb-9 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
    <div className="min-w-0">
      <h1 className="text-[24px] font-semibold leading-tight tracking-[-0.035em]">{title}</h1>
      {description && <p className="mt-1.5 max-w-xl text-sm leading-relaxed text-muted-foreground">{description}</p>}
    </div>
    {action && <div className="shrink-0">{action}</div>}
  </div>;
}
