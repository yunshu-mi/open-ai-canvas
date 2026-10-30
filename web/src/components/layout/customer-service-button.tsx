import { useState } from "react";
import { Headset } from "lucide-react";
import { AppModal } from "@/components/ui/product/app-modal";
import { useAppearanceStore } from "@/stores/use-appearance-store";
import { CustomerServiceContent } from "./customer-service-content";

export function CustomerServiceButton() {
    const value = useAppearanceStore((state) => state.appearance.customerService);
    const [open, setOpen] = useState(false);
    if (!value?.enabled) return null;

    return (
        <>
            <button type="button" className="app-workspace-topbar-feedback-pill" aria-label="联系客服" aria-haspopup="dialog" aria-expanded={open} onClick={() => setOpen(true)}>
                <Headset aria-hidden="true" />
                <span>客服</span>
            </button>
            <AppModal open={open} title={value.title} footer={null} width={460} centered onCancel={() => setOpen(false)}>
                <CustomerServiceContent value={value} />
            </AppModal>
        </>
    );
}
