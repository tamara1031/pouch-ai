import type { FieldSchema } from "../../types";

interface Props {
    id: string;
    schema: FieldSchema;
    value: any;
    onUpdate: (value: string, type: string) => void;
}

export default function SchemaField({ id, schema, value, onUpdate }: Props) {
    return (
        <div class="form-control">
            <label class="label pt-0"><span class="label-text text-[10px] text-white/40 uppercase font-semibold">{schema.displayName || id.replace(/_/g, " ")}</span></label>
            {schema.type === "select" ? (
                <select
                    class="select select-bordered select-xs bg-white/5 border-white/5 rounded-lg text-[10px]"
                    value={value}
                    onChange={(e) => onUpdate(e.currentTarget.value, schema.type)}
                >
                    {schema.options?.map(opt => <option value={opt} key={opt}>{opt}</option>)}
                </select>
            ) : (
                <input
                    type={schema.type === "number" ? "number" : "text"}
                    placeholder={schema.description}
                    value={value}
                    onInput={(e) => onUpdate(e.currentTarget.value, schema.type)}
                    class="input input-bordered input-xs bg-white/5 border-white/5 rounded-lg font-mono text-[10px]"
                />
            )}
        </div>
    );
}
