interface PageHeaderProps {
  title: string;
  description?: string;
  action?: React.ReactNode;
}

export function PageHeader({ title, description, action }: PageHeaderProps) {
  return <div className="mb-7 flex flex-col justify-between gap-4 border-b pb-6 sm:flex-row sm:items-center">
    <div className="min-w-0">
      <h1 className="text-[22px] font-semibold leading-tight tracking-[-0.035em]">{title}</h1>
      {description && <p className="mt-2 max-w-xl text-xs leading-relaxed text-muted-foreground">{description}</p>}
    </div>
    {action && <div className="shrink-0">{action}</div>}
  </div>;
}
