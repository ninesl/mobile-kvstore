-- name: GetBlobRefsByScopeAnd10Refs :many
WITH refs(id, meta_tag) AS (
  SELECT sqlc.narg(id_00) AS id, sqlc.narg(meta_tag_00) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_01) AS id, sqlc.narg(meta_tag_01) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_02) AS id, sqlc.narg(meta_tag_02) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_03) AS id, sqlc.narg(meta_tag_03) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_04) AS id, sqlc.narg(meta_tag_04) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_05) AS id, sqlc.narg(meta_tag_05) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_06) AS id, sqlc.narg(meta_tag_06) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_07) AS id, sqlc.narg(meta_tag_07) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_08) AS id, sqlc.narg(meta_tag_08) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_09) AS id, sqlc.narg(meta_tag_09) AS meta_tag
)
SELECT DISTINCT br.blob_id
FROM refs r
JOIN blob_refs br
  ON br.id = r.id
 AND br.meta_tag = r.meta_tag
WHERE br.namespace = sqlc.arg(namespace)
  AND br.subject = sqlc.arg(subject)
  -- Both checks are intentionally redundant: caller promises refs are all-null or all-non-null pairs.
  AND r.id IS NOT NULL
  AND r.meta_tag IS NOT NULL;

-- name: GetBlobRefsByScopeAnd100Refs :many
WITH refs(id, meta_tag) AS (
  SELECT sqlc.narg(id_00) AS id, sqlc.narg(meta_tag_00) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_01) AS id, sqlc.narg(meta_tag_01) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_02) AS id, sqlc.narg(meta_tag_02) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_03) AS id, sqlc.narg(meta_tag_03) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_04) AS id, sqlc.narg(meta_tag_04) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_05) AS id, sqlc.narg(meta_tag_05) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_06) AS id, sqlc.narg(meta_tag_06) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_07) AS id, sqlc.narg(meta_tag_07) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_08) AS id, sqlc.narg(meta_tag_08) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_09) AS id, sqlc.narg(meta_tag_09) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_10) AS id, sqlc.narg(meta_tag_10) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_11) AS id, sqlc.narg(meta_tag_11) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_12) AS id, sqlc.narg(meta_tag_12) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_13) AS id, sqlc.narg(meta_tag_13) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_14) AS id, sqlc.narg(meta_tag_14) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_15) AS id, sqlc.narg(meta_tag_15) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_16) AS id, sqlc.narg(meta_tag_16) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_17) AS id, sqlc.narg(meta_tag_17) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_18) AS id, sqlc.narg(meta_tag_18) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_19) AS id, sqlc.narg(meta_tag_19) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_20) AS id, sqlc.narg(meta_tag_20) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_21) AS id, sqlc.narg(meta_tag_21) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_22) AS id, sqlc.narg(meta_tag_22) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_23) AS id, sqlc.narg(meta_tag_23) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_24) AS id, sqlc.narg(meta_tag_24) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_25) AS id, sqlc.narg(meta_tag_25) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_26) AS id, sqlc.narg(meta_tag_26) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_27) AS id, sqlc.narg(meta_tag_27) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_28) AS id, sqlc.narg(meta_tag_28) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_29) AS id, sqlc.narg(meta_tag_29) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_30) AS id, sqlc.narg(meta_tag_30) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_31) AS id, sqlc.narg(meta_tag_31) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_32) AS id, sqlc.narg(meta_tag_32) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_33) AS id, sqlc.narg(meta_tag_33) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_34) AS id, sqlc.narg(meta_tag_34) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_35) AS id, sqlc.narg(meta_tag_35) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_36) AS id, sqlc.narg(meta_tag_36) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_37) AS id, sqlc.narg(meta_tag_37) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_38) AS id, sqlc.narg(meta_tag_38) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_39) AS id, sqlc.narg(meta_tag_39) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_40) AS id, sqlc.narg(meta_tag_40) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_41) AS id, sqlc.narg(meta_tag_41) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_42) AS id, sqlc.narg(meta_tag_42) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_43) AS id, sqlc.narg(meta_tag_43) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_44) AS id, sqlc.narg(meta_tag_44) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_45) AS id, sqlc.narg(meta_tag_45) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_46) AS id, sqlc.narg(meta_tag_46) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_47) AS id, sqlc.narg(meta_tag_47) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_48) AS id, sqlc.narg(meta_tag_48) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_49) AS id, sqlc.narg(meta_tag_49) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_50) AS id, sqlc.narg(meta_tag_50) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_51) AS id, sqlc.narg(meta_tag_51) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_52) AS id, sqlc.narg(meta_tag_52) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_53) AS id, sqlc.narg(meta_tag_53) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_54) AS id, sqlc.narg(meta_tag_54) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_55) AS id, sqlc.narg(meta_tag_55) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_56) AS id, sqlc.narg(meta_tag_56) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_57) AS id, sqlc.narg(meta_tag_57) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_58) AS id, sqlc.narg(meta_tag_58) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_59) AS id, sqlc.narg(meta_tag_59) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_60) AS id, sqlc.narg(meta_tag_60) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_61) AS id, sqlc.narg(meta_tag_61) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_62) AS id, sqlc.narg(meta_tag_62) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_63) AS id, sqlc.narg(meta_tag_63) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_64) AS id, sqlc.narg(meta_tag_64) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_65) AS id, sqlc.narg(meta_tag_65) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_66) AS id, sqlc.narg(meta_tag_66) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_67) AS id, sqlc.narg(meta_tag_67) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_68) AS id, sqlc.narg(meta_tag_68) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_69) AS id, sqlc.narg(meta_tag_69) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_70) AS id, sqlc.narg(meta_tag_70) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_71) AS id, sqlc.narg(meta_tag_71) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_72) AS id, sqlc.narg(meta_tag_72) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_73) AS id, sqlc.narg(meta_tag_73) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_74) AS id, sqlc.narg(meta_tag_74) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_75) AS id, sqlc.narg(meta_tag_75) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_76) AS id, sqlc.narg(meta_tag_76) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_77) AS id, sqlc.narg(meta_tag_77) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_78) AS id, sqlc.narg(meta_tag_78) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_79) AS id, sqlc.narg(meta_tag_79) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_80) AS id, sqlc.narg(meta_tag_80) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_81) AS id, sqlc.narg(meta_tag_81) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_82) AS id, sqlc.narg(meta_tag_82) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_83) AS id, sqlc.narg(meta_tag_83) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_84) AS id, sqlc.narg(meta_tag_84) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_85) AS id, sqlc.narg(meta_tag_85) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_86) AS id, sqlc.narg(meta_tag_86) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_87) AS id, sqlc.narg(meta_tag_87) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_88) AS id, sqlc.narg(meta_tag_88) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_89) AS id, sqlc.narg(meta_tag_89) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_90) AS id, sqlc.narg(meta_tag_90) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_91) AS id, sqlc.narg(meta_tag_91) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_92) AS id, sqlc.narg(meta_tag_92) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_93) AS id, sqlc.narg(meta_tag_93) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_94) AS id, sqlc.narg(meta_tag_94) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_95) AS id, sqlc.narg(meta_tag_95) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_96) AS id, sqlc.narg(meta_tag_96) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_97) AS id, sqlc.narg(meta_tag_97) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_98) AS id, sqlc.narg(meta_tag_98) AS meta_tag
  UNION ALL SELECT sqlc.narg(id_99) AS id, sqlc.narg(meta_tag_99) AS meta_tag
)
SELECT DISTINCT br.blob_id
FROM refs r
JOIN blob_refs br
  ON br.id = r.id
 AND br.meta_tag = r.meta_tag
WHERE br.namespace = sqlc.arg(namespace)
  AND br.subject = sqlc.arg(subject)
  -- Both checks are intentionally redundant: caller promises refs are all-null or all-non-null pairs.
  AND r.id IS NOT NULL
  AND r.meta_tag IS NOT NULL;
