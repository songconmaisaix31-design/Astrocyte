import { useRoute } from './router';
import { Nav } from './components/Nav';
import { FixtureBanner } from './components/FixtureBanner';
import { AttentionPage } from './pages/attention/AttentionPage';
import { WorkspacePage } from './pages/workspace/WorkspacePage';
import { SwarmPage } from './pages/swarm/SwarmPage';
import { isFixtureMode } from './fixtures';

export function App() {
  const route = useRoute();
  const fixture = isFixtureMode(route.query);
  const path = route.path;

  let page: React.ReactNode;
  if (path.startsWith('/attention')) {
    page = <AttentionPage fixture={fixture} />;
  } else if (path.startsWith('/workspace')) {
    page = <WorkspacePage fixture={fixture} />;
  } else if (path.startsWith('/swarm')) {
    page = <SwarmPage fixture={fixture} />;
  } else {
    page = <AttentionPage fixture={fixture} />;
  }

  return (
    <>
      <Nav currentPath={path} />
      <main
        style={{
          flex: 1,
          marginLeft: 'var(--sidebar-width)',
          padding: 'var(--space-6)',
          maxWidth: 'var(--content-max-width)',
        }}
      >
        {fixture && <FixtureBanner />}
        {page}
      </main>
    </>
  );
}
