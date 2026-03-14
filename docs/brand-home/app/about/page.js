import HomeShell from '../../components/home-shell';
import AboutRootLayout from '../../components/about-root-layout';

export default function AboutPage() {
  return (
    <HomeShell preset="architect" surface="about" headerTitle="about / root">
      <AboutRootLayout />
    </HomeShell>
  );
}
