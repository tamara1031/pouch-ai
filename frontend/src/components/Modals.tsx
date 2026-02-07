import { useEffect } from "preact/hooks";
import { usePluginInfo } from "../hooks/usePluginInfo";
import CreateKeyModal from "./modals/CreateKeyModal";
import EditKeyModal from "./modals/EditKeyModal";
import NewKeyDisplayModal from "./modals/NewKeyDisplayModal";
import {
    isCreateModalOpen,
    isEditModalOpen,
    isNewKeyModalOpen,
    editKeySignal,
    newKeyRawSignal,
    openNewKeyModal,
    openEditModal
} from "../hooks/useModalState";

export default function Modals() {
    const { middlewareInfo, providerInfo } = usePluginInfo();

    useEffect(() => {
        const handleOpenCreate = () => isCreateModalOpen.value = true;
        const handleOpenEdit = (e: any) => openEditModal(e.detail);

        window.addEventListener('open-create-modal', handleOpenCreate);
        window.addEventListener('open-edit-modal', handleOpenEdit);

        return () => {
            window.removeEventListener('open-create-modal', handleOpenCreate);
            window.removeEventListener('open-edit-modal', handleOpenEdit);
        };
    }, []);

    const handleCreateSuccess = (rawKey: string) => {
        isCreateModalOpen.value = false;
        openNewKeyModal(rawKey);
    };

    return (
        <>
            {isCreateModalOpen.value && (
                <CreateKeyModal
                    isOpen={isCreateModalOpen.value}
                    onClose={() => isCreateModalOpen.value = false}
                    onSuccess={handleCreateSuccess}
                    middlewareInfos={middlewareInfo}
                    providerInfos={providerInfo}
                />
            )}
            {isEditModalOpen.value && (
                <EditKeyModal
                    isOpen={isEditModalOpen.value}
                    onClose={() => isEditModalOpen.value = false}
                    editKey={editKeySignal.value}
                    middlewareInfos={middlewareInfo}
                    providerInfos={providerInfo}
                />
            )}
            {isNewKeyModalOpen.value && (
                <NewKeyDisplayModal
                    isOpen={isNewKeyModalOpen.value}
                    onClose={() => isNewKeyModalOpen.value = false}
                    newKeyRaw={newKeyRawSignal.value}
                />
            )}
        </>
    );
}
