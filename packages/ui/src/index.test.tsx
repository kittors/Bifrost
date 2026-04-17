import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";

import { Badge, Button, cn, EmptyState, ErrorState, Input, Table, Tabs } from "./index";

describe("@bifrost/ui", () => {
  it("renders button and input with compact design token classes", () => {
    const buttonMarkup = renderToStaticMarkup(
      <Button size="md" variant="primary">
        Open GitLab
      </Button>,
    );
    const inputMarkup = renderToStaticMarkup(<Input placeholder="Server URL" size="lg" />);

    expect(buttonMarkup).toContain('data-slot="button"');
    expect(buttonMarkup).toContain("h-[36px]");
    expect(buttonMarkup).toContain("bg-brand");
    expect(inputMarkup).toContain('data-slot="input"');
    expect(inputMarkup).toContain("h-[40px]");
    expect(inputMarkup).toContain("rounded-[6px]");
  });

  it("renders empty and error states with requestId support", () => {
    const emptyMarkup = renderToStaticMarkup(
      <EmptyState title="No Services" description="No service is available yet." />,
    );
    const errorMarkup = renderToStaticMarkup(
      <ErrorState
        title="Access Denied"
        description="You do not have access to this service."
        requestId="req_01JZERROR"
      />,
    );

    expect(emptyMarkup).toContain("No Services");
    expect(errorMarkup).toContain("req_01JZERROR");
    expect(errorMarkup).toContain("Request ID");
  });

  it("exports shared composition primitives", () => {
    expect(typeof Badge).toBe("function");
    expect(typeof Table.Root).toBe("function");
    expect(typeof Tabs.Root).toBe("function");
    expect(cn("bg-brand", undefined, "text-text-primary")).toBe("bg-brand text-text-primary");
  });
});
