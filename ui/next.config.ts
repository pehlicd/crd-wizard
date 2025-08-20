import type {NextConfig} from 'next';

const nextConfig: NextConfig = {
  distDir: 'dist',
  output: 'export',
  typescript: {
    ignoreBuildErrors: true,
  },
  eslint: {
    ignoreDuringBuilds: true,
  },
};

export default nextConfig;
