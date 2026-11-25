interface PageHeaderProps {
  title: string;
  description?: React.ReactNode;
}

export function PageHeader({ title, description }: PageHeaderProps) {
  return (
    <header className="space-y-2 border-b pb-6">
      <h1 className="text-[28px] font-semibold leading-tight tracking-tight sm:text-[32px]">{title}</h1>
      {description && <p className="text-sm leading-6 text-muted-foreground">{description}</p>}
    </header>
  );
}
