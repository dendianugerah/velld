import React from 'react';
import Image from 'next/image';

const Logo: React.FC = () => {
  return (
    <div className="flex items-center gap-3">
      <Image
        src="/logo.png"
        alt="Velld Logo"
        width={36}
        height={36}
        className="rounded-lg"
      />
      <span className="text-base font-semibold bg-clip-text text-transparent bg-gradient-to-r from-blue-500 to-blue-600">
        Velld
      </span>
    </div>
  );
};

export default Logo;