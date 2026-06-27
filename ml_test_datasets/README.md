# 🤖 ML Test Datasets — VisionPrice Proveedores Backend

Colección de **22 datasets sintéticos** para probar los tres algoritmos de Machine Learning implementados en el backend.

---

## Algoritmos Cubiertos

| Algoritmo | Use Case | Umbral |
|-----------|----------|--------|
| **Isolation Forest** | Detección de anomalías en precios | Score ≥ `0.65` → anomalía |
| **TF-IDF + Cosine Similarity** | Detección de productos duplicados | Score ≥ `0.85` → duplicado |
| **TF-IDF + Cosine Similarity** | Clasificación de productos vs catálogo maestro | Busca mayor score |

---

## Estructura de Carpetas

```
ml_test_datasets/
├── README.md
├── anomaly_detection/          # 7 datasets — Isolation Forest
├── duplicate_detection/        # 6 datasets — TF-IDF duplicados
├── classification/             # 4 datasets — TF-IDF clasificador
└── edge_cases/                 # 5 datasets — Casos límite
```

---

## Descripción de Datasets

### 📊 Detección de Anomalías (Isolation Forest)
Requiere al menos 10 precios de mercado por par `categoria|unidad`.

| Dataset | Descripción | Anomalías Esperadas |
|---------|-------------|:-------------------:|
| dataset_01 | Construcción, precios normales | 0 |
| dataset_02 | Construcción con anomalías (5-10× el promedio) | 3–4 |
| dataset_03 | Eléctrico con anomalías | 2–3 |
| dataset_04 | Plomería mixto | 2 |
| dataset_05 | Herramientas con precios extremos | 4–5 |
| dataset_06 | Pinturas con anomalías sutiles | 1–2 |
| dataset_07 | Aceros con anomalías en kg | 2–3 |

### 🔍 Detección de Duplicados (TF-IDF + Cosine Similarity)
Umbral: 0.85. El procesamiento paralelo se activa con más de 200 productos.

| Dataset | Descripción | Pares Esperados |
|---------|-------------|:---------------:|
| dataset_08 | Duplicados obvios | 4–5 pares |
| dataset_09 | Duplicados por abreviaciones | 3–4 pares |
| dataset_10 | Duplicados con typos | 2–3 pares |
| dataset_11 | Sin duplicados | 0 pares |
| dataset_12 | Múltiples grupos de duplicados | 6–8 pares |
| dataset_13 | 220+ productos (activa procesamiento paralelo) | Variable |

### 🏷️ Clasificación de Productos (TF-IDF Classifier)

| Dataset | Descripción | Confianza Esperada |
|---------|-------------|:------------------:|
| dataset_14 | Catálogo maestro estándar (30 productos) | — |
| dataset_15 | Productos bien nombrados | Alta (> 0.7) |
| dataset_16 | Productos ambiguos | Media (0.3–0.6) |
| dataset_17 | Productos sin match en catálogo | Baja (< 0.2) |

### ⚠️ Casos Límite (Edge Cases)

| Dataset | Descripción | Propósito |
|---------|-------------|-----------|
| dataset_18 | Solo 2 productos | Verifica condición mínima |
| dataset_19 | Precios en frontera umbral 0.65 | Comportamiento en el límite |
| dataset_20 | Nombres con acentos, ñ y símbolos | Prueba el tokenizador |
| dataset_21 | 500+ productos | Estrés y rendimiento paralelo |
| dataset_22 | Escenario mixto completo | Integración end-to-end |

---

## Formato JSON de los Datasets

```json
{
  "description": "Descripción del dataset",
  "algorithm": "isolation_forest | tfidf_cosine_duplicates | tfidf_cosine_classifier | all",
  "expected_behavior": "Qué se espera que detecte el algoritmo",
  "provider_id": "UUID del proveedor de prueba",
  "market_prices_by_category_unit": {
    "Categoria|unidad": [10.0, 11.0, 12.0]
  },
  "products": [
    { "id": "UUID", "name": "Nombre", "category": "Cat", "unit": "ud", "price": 0.0 }
  ],
  "standard_catalog": [
    { "id": 1, "name": "Nombre estándar", "category": "Categoría" }
  ]
}
```

---

## Notas Importantes

- Precios en **Quetzales guatemaltecos (Q)**.
- Isolation Forest: necesita mínimo 10 precios de mercado por `categoria|unidad`.
- Duplicados paralelo: se activa con más de 200 productos (`parallelMinProducts = 200`).
- Isolation Forest: 100 árboles, submuestra hasta 256 datos.

---
*Generado para testing de ML — VisionPrice Proveedores Backend*
