import type { Metadata } from "next";

import { PageHeader } from "@/components/shared/page-header";
import { DocumentList } from "@/features/knowledge/components/document-list";
import { IngestDocumentForm } from "@/features/knowledge/components/ingest-document-form";
import { KnowledgeSearch } from "@/features/knowledge/components/knowledge-search";

export const metadata: Metadata = { title: "Knowledge | NovaFlow" };

export default function KnowledgePage() {
  return (
    <>
      <PageHeader
        title="Knowledge"
        description="The documents NovaFlow retrieves from when it writes captions for you."
      />

      <div className="flex flex-col gap-8">
        <KnowledgeSearch />

        <section className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <h2 className="text-lg font-semibold tracking-tight">Documents</h2>
            <p className="text-sm text-muted-foreground">
              Everything indexed for retrieval. Documents are chunked and embedded after they are
              added.
            </p>
          </div>

          <IngestDocumentForm />
          <DocumentList />
        </section>
      </div>
    </>
  );
}
