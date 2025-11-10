/* Scripts JavaScript du service de monitoring
 * Projet de session A25
 * By : Leandre Kanmegne
 *
 * Gère l'interface utilisateur et les appels API
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
const SEUIL_LENTE_MS = 800;
let intervalId = null;
let enCours = false;

// affiche un message à l'utilisateur
function afficherMessage(texte, type = 'info') {
  zoneMessage.textContent = texte || '';
  zoneMessage.className = `message ${type}`;
}

// vide la console des résultats
function viderConsole() {
  liste.innerHTML = '';
  if (!document.getElementById('ligne-vide')) {
    const vide = document.createElement('div');
    vide.id = 'ligne-vide';
    vide.className = 'vide';
    vide.textContent = "Aucun résultat pour l'instant.";
    liste.parentElement.insertBefore(vide, liste);
  }
}

// formate l'heure pour l'affichage
function formaterHeure(dateISO) {
  try {
    const date = new Date(dateISO);
    if (isNaN(date.getTime())) throw new Error('date invalide');
    return date.toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  } catch {
    return '—';
  }
}

// crée une ligne de résultat dans la console
function creerLigne(statut) {
  // supprime la ligne vide si c'est le premier résultat
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

  // détermine le statut visuel
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

  // ajoute le message d'erreur si présent
  if (!est_disponible && message_erreur) {
    const sous = document.createElement('div');
    sous.className = 'sous-ligne';
    sous.textContent = message_erreur;
    ligne.appendChild(sous);
  }

  // ajoute en haut de la liste
  if (liste.firstChild) {
    liste.insertBefore(ligne, liste.firstChild);
  } else {
    liste.appendChild(ligne);
  }
}

// fait un appel à l'API
async function appelAPI(url, options) {
  try {
    const resp = await fetch(url, options);
    const json = await resp.json().catch(() => ({}));
    if (!resp.ok) {
      const msg = json?.error || json?.message || `Erreur HTTP ${resp.status}`;
      throw new Error(msg);
    }
    return json;
  } catch (err) {
    throw new Error(err.message || 'Erreur réseau');
  }
}

// vérifie une URL via l'API
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

  const statut = reponse.statut;
  if (!statut) throw new Error('Réponse API invalide');

  creerLigne(statut);
  return statut;
}

// charge les derniers résultats depuis l'API
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

  viderConsole();
  for (const statut of listeResultats) {
    creerLigne(statut);
  }

  return listeResultats.length;
}

// soumission du formulaire
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

// bouton vider
btnVider.addEventListener('click', async () => {
  try {
    await appelAPI('/api/resultats', { method: 'DELETE' });
    viderConsole();
    afficherMessage('Base et console vidées.', 'ok');
  } catch (err) {
    afficherMessage(`Échec de la suppression: ${err.message}`, 'err');
  }
});

// chargement initial des résultats
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

// démarrer l'auto-ping
function startAutoPing() {
  const url = (champURL.value || '').trim();
  const secondes = parseInt(champFreq.value, 10);

  if (!url) {
    afficherMessage("Saisissez d'abord une URL.", 'info');
    champURL.focus();
    return;
  }
  if (!Number.isFinite(secondes) || secondes < 1) {
    afficherMessage('Fréquence invalide. Entrez un nombre ≥ 1.', 'err');
    champFreq.focus();
    return;
  }

  if (intervalId) clearInterval(intervalId);

  afficherMessage(`Auto-ping activé: toutes les ${secondes}s`, 'ok');

  intervalId = setInterval(async () => {
    if (enCours) return;
    enCours = true;
    try {
      await verifierUneURL(url);
    } catch (err) {
      afficherMessage(err.message, 'err');
    } finally {
      enCours = false;
    }
  }, secondes * 1000);
}

// arrêter l'auto-ping
function stopAutoPing() {
  if (intervalId) {
    clearInterval(intervalId);
    intervalId = null;
    afficherMessage('Auto-ping arrêté.', 'info');
  }
}

btnAuto?.addEventListener('click', startAutoPing);
btnStop?.addEventListener('click', stopAutoPing);
