import type { FieldSchema } from "../../../../types";

interface Props {
    id: string;
    schema: FieldSchema;
    value: any;
    onUpdate: (value: string, type: string) => void;
}

export default function SchemaField({ id, schema, value, onUpdate }: Props) {
    const displayName = schema.displayName || id.replace(/_/g, " ");

    const handleChange = (e: any) => {
        onUpdate(e.currentTarget.value, schema.type);
    };

    return (
        <div class="form-control w-full">
            <label class="label pb-1">
                <span class="label-text text-xs text-white/50 font-medium capitalize">
                    {displayName}
                </span>
            </label>
            {schema.type === "select" ? (
                <select
                    class="select select-bordered w-full bg-base-200/50 border-white/10 rounded-lg text-sm h-10"
                    value={value}
                    onChange={handleChange}
                >
                    {schema.options?.map(opt => <option value={opt} key={opt}>{opt}</option>)}
                </select>
            ) : schema.type === "number" ? (
                <input
                    type="number"
                    placeholder={schema.description}
                    value={value}
                    onInput={handleChange}
                    class="input input-bordered w-full bg-base-200/50 border-white/10 rounded-lg text-sm h-10"
                />
            ) : schema.type === "boolean" ? (
                 <select
                    class="select select-bordered w-full bg-base-200/50 border-white/10 rounded-lg text-sm h-10"
                    value={String(value)}
                    onChange={handleChange}
                >
                    <option value="true">True</option>
                    <option value="false">False</option>
                </select>
            ) : (
                <textarea
                    placeholder={schema.description}
                    value={value}
                    onInput={handleChange}
                    class="textarea textarea-bordered w-full bg-base-200/50 border-white/10 rounded-lg text-sm min-h-20 resize-y"
                    rows={3}
                />
            )}
            {schema.description && (
                <label class="label pt-1">
                    <span class="label-text-alt text-white/30 text-xs">{schema.description}</span>
                </label>
            )}
        </div>
    );
}
