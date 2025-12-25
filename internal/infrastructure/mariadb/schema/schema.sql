-- Create "vm_templates" table
CREATE TABLE `vm_templates` (
    `vmid` int unsigned NOT NULL,
    `name` varchar(100) NOT NULL,
    PRIMARY KEY (`vmid`)
) COLLATE utf8mb4_uca1400_ai_ci;

-- Create "instances" table
CREATE TABLE `instances` (
    `vmid` int unsigned NOT NULL,
    `user_id` varchar(100) NOT NULL,
    `template_vmid` int unsigned NOT NULL,
    `ip_address` varchar(45) NULL,
    `created_at` TIMESTAMP NULL DEFAULT (current_timestamp()),
    PRIMARY KEY (`vmid`),
    INDEX `instances_vm_templates_FK` (`template_vmid`),
    CONSTRAINT `instances_vm_templates_FK`
    FOREIGN KEY (`template_vmid`) REFERENCES `vm_templates` (`vmid`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) COLLATE utf8mb4_uca1400_ai_ci;
