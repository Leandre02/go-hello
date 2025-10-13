/* Les scripts JavaScript de la page d'accueil de mon service de monitoring */
/* Script front du MVP de monitoring
 * - Envoie POST /api/check {url}
 * - Récupère GET /api/resultats?limit=50
 * - Affiche une “console” de pings (les plus récents en haut)
 * Noms FR + commentaires explicites.
 */

const champURL = document.getElementById('champ-url');
const formulaire = document.getElementById('formulaire');
const btnTester = document.getElementById('btn-tester');
const btnVider = document.getElementById('btn-vider');
const champFreq = document.getElementById('champ-freq');
const btnAuto = document.getElementById('btn-auto');
const btnStop = document.getElementById('btn-stop');
const zoneMessage = document.getElementById('zone-message');
const liste = document.getElementById('liste');
const ligneVide = document.getElementById('ligne-vide');

const LIMITE_PAR_DEFAUT = 50;
const SEUIL_LENTE_MS = 800; // doit être cohérent avec le backend (variable d’env. SEUIL_LENTE_MS)
let intervalId = null;
let enCours = false; // anti-chevauchement

// -------- Utilitaires UI --------

function afficherMessage(texte, type = 'info') {
  zoneMessage.textContent = texte || '';
  zoneMessage.className = `message ${type}`;
}

function viderConsole() {
  liste.innerHTML = '';
  if (!document.getElementById('ligne-vide')) {
    const v = document.createElement('div');
    v.id = 'ligne-vide';
    v.className = 'vide';
    v.textContent = 'Aucun résultat pour l’instant.';
    liste.parentElement.insertBefore(v, liste);
  }
}

function formaterHeure(dateISO) {
  try {
    const d = new Date(dateISO);
    if (isNaN(d.getTime())) throw new Error('invalid date');
    return d.toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  } catch {
    return '—';
  }
}

function creerLigne(statut) {
  // Supprime la ligne “vide” si c’est le premier résultat
  const vide = document.getElementById('ligne-vide');
  if (vide) vide.remove();

  const {
    url,
    est_disponible,
    code_http,
    message_erreur,
    latence_ms,
    verifie_a,
  } = statut;

  // Détermine le statut visuel (ok / warn / err)
  let texteStatut = 'EN LIGNE';
  let classeStatut = 'ok';
  if (!est_disponible) {
    texteStatut = 'HORS SERVICE';
    classeStatut = 'err';
  } else if (latence_ms >= SEUIL_LENTE_MS) {
    texteStatut = 'LENT';
    classeStatut = 'warn';
  }

  const ligne = document.createElement('div');
  ligne.className = 'ligne';

  ligne.innerHTML = `
    <div>${formaterHeure(verifie_a)}</div>
    <div class="url" title="${url}">${url}</div>
    <div class="statut ${classeStatut}">${texteStatut}</div>
    <div class="badge">${latence_ms ?? '—'} ms</div>
    <div class="badge">${code_http ?? '—'}</div>
  `;

  if (!est_disponible && message_erreur) {
    // Ajoute une sous-ligne compacte pour le message d’erreur
    const sous = document.createElement('div');
    sous.className = 'sous-ligne';
    sous.textContent = message_erreur;
    ligne.appendChild(sous);
  }

  // Ajoute en tête (les plus récents en haut)
  if (liste.firstChild) {
    liste.insertBefore(ligne, liste.firstChild);
  } else {
    liste.appendChild(ligne);
  }
}

// -------- Appels API --------

async function appelAPI(url, options) {
  try {
    const rsp = await fetch(url, options);
    const json = await rsp.json().catch(() => ({}));
    if (!rsp.ok) {
      const msg = json?.error || json?.message || `Erreur HTTP ${rsp.status}`;
      throw new Error(msg);
    }
    return json;
  } catch (err) {
    throw new Error(err.message || 'Erreur réseau');
  }
}

async function verifierUneURL(url) {
  if (!url || !/^https?:\/\//i.test(url)) {
    throw new Error(
      'Veuillez saisir une URL valide commençant par http:// ou https://',
    );
  }
  const reponse = await appelAPI('/api/verifier', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ url }),
  });
  const statut = reponse.statut; // ton backend renvoie toujours {statut:{...}}
  if (!statut) throw new Error('Réponse API invalide');
  creerLigne(statut);
  return statut;
}

async function chargerResultats(limit = LIMITE_PAR_DEFAUT) {
  const reponse = await appelAPI(
    `/api/resultats?limit=${encodeURIComponent(limit)}`,
    {
      method: 'GET',
    },
  );
  const listeResultats = Array.isArray(reponse)
    ? reponse
    : reponse?.resultats ?? [];
  // Réinitialise puis rend les résultats
  viderConsole();
  for (const s of listeResultats) {
    creerLigne(s);
  }
  return listeResultats.length;
}

// -------- Gestion des événements --------

formulaire.addEventListener('submit', async (e) => {
  e.preventDefault();
  const url = (champURL.value || '').trim();
  if (!url) return;

  afficherMessage('Vérification en cours…', 'info');
  btnTester.disabled = true;

  try {
    await verifierUneURL(url);
    afficherMessage('Ping effectué.', 'ok');
    champURL.value = '';
    champURL.focus();
  } catch (err) {
    afficherMessage(err.message, 'err');
  } finally {
    btnTester.disabled = false;
  }
});

btnVider.addEventListener('click', () => {
  viderConsole();
  afficherMessage('Console vidée.', 'info');
});

// Au chargement : récupère les derniers résultats
window.addEventListener('DOMContentLoaded', async () => {
  try {
    await chargerResultats(LIMITE_PAR_DEFAUT);
  } catch (err) {
    afficherMessage(
      `Impossible de charger les résultats : ${err.message}`,
      'err',
    );
  }
});

// -------- Auto-ping --------

function startAutoPing() {
  const url = (champURL.value || '').trim();
  const secs = parseInt(champFreq.value, 10);
  if (!url) {
    afficherMessage("Saisissez d'abord une URL.", 'info');
    champURL.focus();
    return;
  }
  if (!Number.isFinite(secs) || secs < 1) {
    afficherMessage('Fréquence invalide. Entrez un nombre ≥ 1.', 'err');
    champFreq.focus();
    return;
  }
  if (intervalId) clearInterval(intervalId);
  afficherMessage(`Auto-ping activé: toutes les ${secs}s`, 'ok');
  intervalId = setInterval(async () => {
    if (enCours) return; // pas de chevauchement
    enCours = true;
    try {
      await verifierUneURL(url);
    } catch (err) {
      afficherMessage(err.message, 'err');
    } finally {
      enCours = false;
    }
  }, secs * 1000);
}

function stopAutoPing() {
  if (intervalId) {
    clearInterval(intervalId);
    intervalId = null;
    afficherMessage('Auto-ping arrêté.', 'info');
  }
}

btnAuto?.addEventListener('click', startAutoPing);
btnStop?.addEventListener('click', stopAutoPing);
